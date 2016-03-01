# Terphite

This is a toy/experimental Console [Graphite](http://graphite.readthedocs.org/) Browser loosely based on Graphite Composer.

It uses [blessed](https://github.com/chjj/blessed) and [blessed-contrib](https://github.com/yaronn/blessed-contrib) to do all the heavy lifting. *blessed-contrib* is a library for building console dashboards, it provides the tree and graph widgets. *blessed* is the UI toolkit, it has a DOM-like API and is surprisingly easy to work with.

Next steps might be to add the _Graph Options_ and _Apply Function_ features from Composer. And make a dashboard view a la [blessed-graphite](https://github.com/lovehandle/blessed-graphite) that can display and save a grid of graphs.

#### Demo Screencast
![](http://i.imgur.com/l8LbbrG.gif)

##### Reactions to the Screencast :-)
> grubernaut [4:37 PM]
holy shit

> obfuscurity [8:37 AM]
whoa wtf

>obfuscurity [8:37 AM]
that’s better than the real thing lol

### Install

    npm install -g terphite

### Usage

    terphite http://user:pass@your.graphite.com:1234

### Install and run from source

    git clone git@github.com:benwtr/terphite.git
    cd terphite
    npm install
    ./bin/terphite http://your.graphite.com

#### Getting started with this code (for people unfamiliar with CoffeeScript)

The code in `src/` is CoffeeScript, it gets compiled to JavaScript and output to `lib/`. 

To compile the CoffeeScript source:

    cake build

Or watch source for changes and compile when modified:

    cake watch

Or, if you don't like CoffeeScript, just edit the JS directly. :-)


