const bodyParser = require('body-parser');
const express = require('express');
const redis = require('redis');

const app = express();
const client = redis.createClient();

client.on('error', function(error) {
	console.error(error);
});

app.use(bodyParser.json())

app.post('/intake', async (req, res) => {
	console.log(req.body);

	client.set(`${req.body.event}-${req.body.user}-pos`, JSON.stringify(req.body.position));

	res.status(200).send();
});

app.listen(8000, () => console.log('Listening...'));