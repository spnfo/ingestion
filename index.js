const bodyParser = require('body-parser');
const express = require('express');
const redis = require('redis');

const app = express();
const globalClient = redis.createClient();

globalClient.on('error', function(error) {
	console.error(error);
});

app.use(bodyParser.json())

app.post('/intake', async (req, res) => {

	if (Math.abs(req.body.position[0]) < 90) {

		cur_client = redis.createClient();

		cur_client.on('message', (channel, message) => {
			res.status(200).send(message);
			cur_client.quit();
		});

		cur_client.subscribe(`${req.body.event}-${req.body.user}-reply`);

		globalClient.publish(`${req.body.event}-${req.body.user}-pos`, JSON.stringify(req.body.position), err => {
			if (err) { throw err }
		});
		globalClient.set(`${req.body.event}-${req.body.user}-pos`, JSON.stringify(req.body.position), err => {
			if (err) { throw err }
		});
	}

});

app.listen(8000, () => console.log('Listening...'));