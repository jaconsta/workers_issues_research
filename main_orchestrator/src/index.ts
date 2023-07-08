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
};

function celeryConnect() {
  const backendDb = config.redis.dbNumber * 2;
  const urls = {
    backend: `${config.redis.url}/${backendDb}`,
    broker: `${config.redis.url}/${backendDb * 2}`,
  };
  console.log(urls);
  const client = createClient(urls.backend, urls.broker);
  return client;
}

async function celeryTrigger(conn: Client, taskName: string, arg: unknown[]) {
  const task = conn.createTask(taskName);
  console.log(arg, taskName);
  const applyed = task.applyAsync([...arg]);
  return applyed.get();
}

async function create500Users(conn: Client) {
  let batches = [];
  for (let i = 0; i < 500; i++) {
    const firstName = faker.person.firstName();
    const lastName = faker.person.lastName();
    const email = faker.internet.email();

    batches.push(celeryTrigger(conn, config.tasks.createUser, [firstName, lastName, email]));
    if (batches.length >= 100) {
      // todo: fix error:
      //(node:39679) MaxListenersExceededWarning: Possible EventEmitter memory leak detected. 11 ready listeners added to [Redis]. Use emitter.setMaxListeners() to increase limit
      // (Use `node --trace-warnings ...` to show where the warning was created)
      const batchResult = await Promise.all(batches);
      console.log(batchResult);
      batches = [];
    }
  }
}
(async function main() {
  const celeryConn = celeryConnect();
  await create500Users(celeryConn).then(() => console.log('done'));
})();
