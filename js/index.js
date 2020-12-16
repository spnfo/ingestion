const bodyParser = require('body-parser');
const express = require('express');
const redis = require('redis');

const app = express();

app.use(bodyParser.json())

app.post('/intake', async (req, res) => {

	let globalClient = redis.createClient();

	if (Math.abs(req.body.position[0]) < 90) {

		cur_client = redis.createClient();

		cur_client.on('message', (channel, message) => {

			console.log(channel);

			try {
				res.status(200).send(message);
				cur_client.quit();
			} catch {
				// duplicate request...don't send again.
			}

			// cur_client.quit();
		});

		cur_client.subscribe(`${req.body.event}-${req.body.user}-${req.body.req_id}-reply`);

		globalClient.publish(`${req.body.event}-${req.body.user}-pos`, JSON.stringify({
			position: req.body.position, 
			req_id: req.body.req_id
		}), err => {
			if (err) { throw err }
		});

		globalClient.set(`${req.body.event}-${req.body.user}-pos`, JSON.stringify(req.body.position), err => {
			if (err) { throw err }
		});
	}

});

app.listen(8000, () => console.log('Listening...'));