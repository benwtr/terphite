# Terphite

This is a toy/experimental Console [Graphite](http://graphite.readthedocs.org/) Browser loosely based on Graphite Composer.

It uses [blessed](https://github.com/chjj/blessed) and [blessed-contrib](https://github.com/yaronn/blessed-contrib) (line graph and tree) to do all the heavy lifting, both are awesome. I almost did this with [termui](https://github.com/gizak/termui) but the choice of coffee-script or Go really came down to the availability of a _tree_ widget.

This could be taken much, much further. Implementing the features of Kibana or Grafana in a console app is totally feasible.

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

