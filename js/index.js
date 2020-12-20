const bodyParser = require('body-parser');
const express = require('express');
const redisCluster = require('redis-clustr');

let redis = new redisCluster({
	servers: [
		{
			host: '127.0.0.1',
			port: 7000
		},
		{
			host: '127.0.0.1',
			port: 7001
		},
		{
			host: '127.0.0.1',
			port: 7002
		}
	]
});

const app = express();

app.use(bodyParser.json())

app.post('/intake', async (req, res) => {

	console.log('hello');

	if (Math.abs(req.body.position[0]) < 90) {

		redis.on('message', (channel, message) => {

			console.log(channel);

			try {
				res.status(200).send(JSON.parse(message));
			} catch {
				// duplicate request...don't send again.
			}

			// cur_client.quit();
		});

		redis.subscribe(`${req.body.event}-${req.body.user}-${req.body.req_id}-reply`);

		redis.publish(`${req.body.event}-${req.body.user}-pos`, JSON.stringify({
			position: req.body.position, 
			req_id: req.body.req_id
		}), err => {
			if (err) { throw err }
		});

		redis.set(`${req.body.event}-${req.body.user}-pos`, JSON.stringify(req.body.position), err => {
			if (err) { throw err }
		});
	} else {
		res.setHeader('Content-Type', 'application/json');
		res.status(200).send(JSON.stringify({success: true}));
	}

});

app.listen(8000, () => console.log('Listening...'));