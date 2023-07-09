import { createClient } from 'celery-node';
import Client from 'celery-node/dist/app/client';
import { faker } from '@faker-js/faker';

const config = {
  redis: {
    dbNumber: 1,
    ipFamily: 4,
    url: process.env.REDIS_URL ?? 'redis://redis:6379',
  },
  tasks: {
    createUser: 'users.create',
    messageNotify: 'messages.Notify',
  },
  tunning: {
    createUsers: 20, // default: 500,
    batchSize: 5,
    sendGreeting: true,
  },
};

function celeryConnect(dbMultiplier: number) {
  const backendDb = dbMultiplier * 2;
  const urls = {
    backend: `${config.redis.url}/${backendDb}`,
    broker: `${config.redis.url}/${backendDb * 2}`,
  };
  console.log(urls);
  const client = createClient(urls.backend, urls.broker);

  // Go's celery is not yet compatible with Task protocol V2
  client.conf.TASK_PROTOCOL = 1;

  return client;
}

async function celeryTrigger(conn: Client, taskName: string, arg: unknown[]) {
  const task = conn.createTask(taskName);
  console.log(arg, taskName);
  const applyed = task.applyAsync([...arg]);
  return applyed.get();
}

async function sendGreetingsMessages(conn: Client, newUsers: string[][]) {
  let batches = [];

  for (const user of newUsers) {
    const [, , email] = user;
    batches.push(celeryTrigger(conn, config.tasks.messageNotify, [email, faker.lorem.sentence(10)]));
    if (batches.length >= config.tunning.batchSize / 30) {
      const batchResult = await Promise.all(batches);
      console.log(batchResult);
      batches = [];
    }
  }
}

async function create500Users(conn: Client) {
  let batches = [];
  const newUsers = [];
  for (let i = 0; i < config.tunning.createUsers; i++) {
    const firstName = faker.person.firstName();
    const lastName = faker.person.lastName();
    const email = faker.internet.email();

    const newUser = [firstName, lastName, email];
    batches.push(celeryTrigger(conn, config.tasks.createUser, newUser));
    newUsers.push(newUser);
    if (batches.length >= config.tunning.batchSize) {
      // todo: fix error:
      //(node:39679) MaxListenersExceededWarning: Possible EventEmitter memory leak detected. 11 ready listeners added to [Redis]. Use emitter.setMaxListeners() to increase limit
      // (Use `node --trace-warnings ...` to show where the warning was created)
      const batchResult = await Promise.all(batches);
      console.log(batchResult);
      batches = [];
    }
  }
  return newUsers;
}
(async function main() {
  const celeryConn = celeryConnect(config.redis.dbNumber);
  const newUsers = await create500Users(celeryConn);
  celeryConn.disconnect();
  if (config.tunning.sendGreeting) {
    const celery2 = celeryConnect(config.redis.dbNumber + 2);
    await sendGreetingsMessages(celery2, newUsers);
    celery2.disconnect();
  }
  console.log('done');
})();
