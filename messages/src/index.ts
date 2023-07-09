import { createWorker } from 'celery-node';
import Worker from 'celery-node/dist/app/worker';

let globalCounter = 0;
const config = Object.freeze({
  redis: {
    dbNumber: 3, // 1,
    ipFamily: 4,
    url: process.env.REDIS_URL ?? 'redis://redis:6379',
  },
  tasks: {
    createUser: 'users.create',
    messageNotify: 'messages.Notify',
  },
  globalCounter,
});

function celeryConnect() {
  const backendDb = config.redis.dbNumber * 2;
  const urls = {
    backend: `${config.redis.url}/${backendDb}`,
    broker: `${config.redis.url}/${backendDb * 2}`,
  };
  console.log(urls);
  const client = createWorker(urls.backend, urls.broker);
  return client;
}

function registerProcessMessage(conn: Worker, taskName: string) {
  conn.register(taskName, (userId: string, message: string) =>
    console.log(`(${globalCounter++}) Sending a message to ${userId}, message: ${message}`)
  );
}
function voidHandlers(conn: Worker) {
  conn.register(config.tasks.createUser, () => {
    console.log(`\n (${globalCounter++})`, 'Void handler for,', config.tasks.createUser);
    return 'error_not_me';
  });
}

(async function main() {
  const celeryConn = celeryConnect();
  registerProcessMessage(celeryConn, config.tasks.messageNotify);
  voidHandlers(celeryConn);
  celeryConn.start();
})();
