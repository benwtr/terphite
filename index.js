var process  = require('process');
var path     = require('path');
var Terphite = require("./lib/composer.js");

var node_binary = process.argv.shift();
var script_name = path.basename(process.argv.shift());

if (process.argv.length < 1) {
  console.error("Usage: " + script_name + " https://user:pass@yourgraphite.net:4321");
  process.exit(1);
}

var graphite_uri = process.argv.shift();

t = new Terphite(graphite_uri);
t.composer();
