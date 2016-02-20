blessed = require 'blessed'
contrib = require 'blessed-contrib'
request = require 'request'
path    = require 'path'
_       = require 'lodash'

global[k] = v for k,v of require './helpers'


module.exports = class Terphite
  constructor: (@graphite_uri) ->

  composer: ->

    selected_metrics = []
    time_from        = '-1min'
    autorefresh_time = 10 # default autorefresh time in seconds
    max_data_points  = 300

    graphite_uri = @graphite_uri
    autorefresh = 0

    screen = blessed.screen {
      smartCSR: true
      debug: true # F12 to open debug popup
      title: 'Graphite Browser'
      warnings: true
      dockBorders: true
      ignoreDockContrast: true
    }

    layout = blessed.layout {
      parent: screen
      width:  '100%'
      height: '100%'
      border: 'line'
      layout: 'grid'
      style:
        bg: 'black',
        border:
          fg: 'blue'
    }

    tree = contrib.tree {
      parent: layout
      #label: 'Metrics Browser'
      top:  0
      left: 0
      padding: 1
      width:  '25%+1'
      height: '80%-2'
      template:
        lines: true
    }

    functions_tree = contrib.tree {
      parent: screen
      top:  'center'
      left: 'center'
      padding: 1
      width:  '80%'
      height: '50%'
      hidden: true
      template:
        lines: true
    }

    statusString = ->
      [
        "from: #{time_from}"
        "autorefresh: #{if !autorefresh then 'off' else autorefresh + 's'}"
        "maxdatapoints: #{if !max_data_points then 'unlimited' else max_data_points}"
      ].join('  ')

    setStatus = -> status.setContent statusString()

    status = blessed.box {
      parent: layout
      top: 0
      left: 0
      height: 3
      width: '100%'
      # padding: 1
      content: statusString()
      border:
        type: 'line'
        fg: 'blue'
    }
    screen.append(status)

    help_content =  '  [ decrease time 1min, { decrease time 1s\n'
    help_content += '  ] increase time 1min, } increase time 1s\n'
    help_content += '  t set relative "from" time, m metrics list popup, i set autorefresh interval\n'
    help_content += '  a autorefresh toggle, x set max datapoints\n'
    help_content += '  o open in browser, c copy graphite URI to clipboard (iTerm2 only)\n'
    help_content += 'Metrics List Keys:\n'
    help_content += '  C-a append new target, C-d delete selected, <enter> edit selected\n'

    help = blessed.box {
      parent: layout
      top: '80%-1'
      left: 0
      height: '20%+2'
      content: help_content
      border:
        type: 'line'
        fg: 'blue'
    }
    screen.append(help)

    line = contrib.line {
      parent: layout
      #label: 'Graph'
      left: '25%'
      top: 2
      height: '80%-2'
      width: '75%+1'
      showLegend: true
      legend:
        width: 60
      border:
        type: 'line'
        fg: 'blue'
    }
    screen.append(line)

    time_popup = blessed.prompt {
      parent: layout
      left: 'center'
      top: 'center'
      width: '80%'
      height: 8
      keys: true
      mouse: true
      style:
        fg: 'blue'
      border:
        type: 'line'
        fg: 'gray'
    }
    screen.append time_popup

    metrics_popup = blessed.list {
      parent: layout
      hidden: true
      left: 'center'
      top: 'center'
      width: '90%'
      height: 'half'
      padding: 1
      interactive: true
      items: selected_metrics
      mouse: true
      keys: true
      tags: true
      style:
        #fg: 'blue'
        bg: 'blue'
      border:
        type: 'line'
        fg: 'gray'
    }
    screen.append metrics_popup

    target_popup = blessed.prompt {
      parent: layout
      left: 'center'
      top: 'center'
      width: '80%'
      height: 8
      mouse: true
      keys: true
      style:
        fg: 'blue'
      border:
        type: 'line'
        fg: 'gray'
    }
    screen.append target_popup

    loadMetricsTree = ->
      options = {
        uri: "#{graphite_uri}/metrics/index.json"
        json: true
      }
      request options, (err, resp, body) ->
        metrics = {}
        for metric in body
          obj = createObject metric
          _.merge metrics, obj
        tree.setData(
          extended: true
          name: 'metrics'
          path: 'metrics'
          children: metrics
        )
        screen.render()

      createObject = (key, original_key = '') ->
        obj = {}
        parts = key.split('.')
        original_key = key if original_key == ''
        original_parts = original_key.split('.')
        path = original_parts[0..original_parts.length-parts.length].join('.')
        if (parts.length == 1)
          # leaf
          obj[parts[0]] =
            name: parts[0]
            extended: false
            path: path
            leaf: true
        else if(parts.length > 1)
          remainingParts = parts.slice(1,parts.length).join('.')
          obj[parts[0]] =
            name: parts[0]
            extended: false
            path: path
            leaf: false
            children: createObject(remainingParts, original_key)
        return obj

    metricURI = (metrics, from, params = {}, opts = {format:'json'}) ->
      extra = ''
      extra += '&format=json' if opts.format == 'json'
      maxdp = parseInt(opts.maxDataPoints,10)
      extra += "&maxDataPoints=#{maxdp}" if maxdp > 0
      target =  ''
      target += '&target=' + metrics.join('&target=') if metrics.length
      "#{graphite_uri}/render?from=#{from}#{target}#{extra}"

    fetchMetricData = (metrics, from) ->
      options =
        uri: metricURI metrics, from, {}, {
          format: 'json'
          maxDataPoints: max_data_points
        }
        json: true
      request options, (err, resp, body) ->
        series = for target,i in body
          title: target.target || '[unnamed_target]'
          y: (point[0] for point in target.datapoints)
          x: for point in target.datapoints
            ts = new Date(point[1]*1000)
            "#{ts.getHours()}:#{format_twodigit(ts.getUTCMinutes())}"
          style: { line: colors[i%15] }
        if series.length == 0
          series = [{ title: 'no data', x: [], y: [] }]
        line.setData(series)
        screen.render()

    tree.on 'select', (node) ->
      path = node.path
      if path in selected_metrics
        selected_metrics = selected_metrics.filter (metric) -> metric isnt path
      else if node.leaf
        selected_metrics.push path
      fetchMetricData selected_metrics, time_from

    screen.key ['['], (ch, key) ->
      t = ( parse_time_to_i time_from ) - MIN
      time_from = if t < MIN then '-1min' else seconds_to_time_string t
      setStatus()
      fetchMetricData(selected_metrics, time_from)

    screen.key [']'], (ch, key) ->
      t = ( parse_time_to_i time_from ) + MIN
      time_from = seconds_to_time_string t
      setStatus()
      fetchMetricData(selected_metrics, time_from)

    screen.key ['{'], (ch, key) ->
      t = ( parse_time_to_i time_from ) - H
      time_from = if t < H then '-1h' else seconds_to_time_string t
      setStatus()
      fetchMetricData(selected_metrics, time_from)

    screen.key ['}'], (ch, key) ->
      t = ( parse_time_to_i time_from ) + H
      time_from = seconds_to_time_string t
      setStatus()
      fetchMetricData(selected_metrics, time_from)

    screen.key ['t'], (ch, key) ->
      time_popup.input(
        'set relative _from_ time (y, mon, w, d, h, min, s)\n  eg: -1d12h',
        time_from,
        (err, value) ->
          [ y, mon, w, d, h, min, s ] = parse_time value
          time_from = get_time_string y, mon, w, d, h, min, s
          setStatus()
          fetchMetricData selected_metrics, time_from
      )

    screen.key ['m'], (ch, key) ->
      metrics_popup.setItems selected_metrics
      metrics_popup.focus()
      metrics_popup.show()
      screen.render()

    metrics_popup.on 'select', (item, select) ->
      target_popup.input(
        'Edit target',
        item.getText(),
        (err, value) ->
          selected_metrics[select] = value
          metrics_popup.setItems selected_metrics
          screen.render()
      )

    metrics_popup.key 'C-d', (ch, key) ->
      selected_metrics.splice(metrics_popup.selected, 1)
      metrics_popup.setItems selected_metrics
      fetchMetricData selected_metrics, time_from

    metrics_popup.key 'C-a', (ch, key) ->
      target_popup.input(
        'Add target',
        '',
        (err, value) ->
          selected_metrics.push value
          metrics_popup.setItems selected_metrics
          fetchMetricData selected_metrics, time_from
      )

    metrics_popup.key ['escape'], (ch, key) ->
      metrics_popup.hide()
      tree.focus()
      fetchMetricData selected_metrics, time_from

    # screen.key ['f'], (ch, key) ->
    #   functions_tree.setData(
    #     extended: true
    #     name: 'foo'
    #     children:
    #       'bar':
    #         extended: false
    #         name: 'bar'
    #       'stuff':
    #         extended: false
    #         name: 'stuff'
    #   )
    #   functions_tree.focus()
    #   functions_tree.setFront()
    #   functions_tree.show()
    #   screen.render()

    screen.key ['c'], (ch, key) ->
      screen.cursorReset()
      screen.copyToClipboard(metricURI selected_metrics, time_from, {}, {})
      screen.realloc()
      screen.render()

    screen.key ['o'], (ch, key) ->
      screen.exec 'open', [(metricURI selected_metrics, time_from, {}, {})]

    autorefresh_loop = null

    toggleAutorefresh = ->
      if !autorefresh
        autorefresh = autorefresh_time
        setStatus()
        screen.render()
        autorefresh_loop = setInterval( ->
          fetchMetricData selected_metrics, time_from
        , autorefresh * 1000)
      else
        autorefresh = 0
        clearInterval autorefresh_loop
        setStatus()
        screen.render()
        autorefresh_loop = null

    screen.key ['a'], (ch, key) ->
      toggleAutorefresh()

    screen.key ['i'], (ch, key) ->
      time_popup.input(
        'set autorefresh interval time (seconds)',
        autorefresh_time.toString(),
        (e, v) ->
          t = parseInt(v)
          autorefresh_time = if t > 0 then t else 1
          if autorefresh
            clearInterval autorefresh_loop
          autorefresh = autorefresh_time
          setStatus()
          screen.render()
          autorefresh_loop = setInterval( ->
            fetchMetricData selected_metrics, time_from
          , autorefresh * 1000)
      )

    screen.key ['x'], (ch, key) ->
      time_popup.input(
        'set max datapoints for graphite api to return in json response\n  0 = unlimited',
        max_data_points.toString(),
        (e,v) ->
          p = parseInt(v)
          max_data_points = if p < 0 then 0 else p
          setStatus()
          fetchMetricData selected_metrics, time_from
      )

    screen.key(['q', 'C-c'], (ch, key) ->
      process.exit(0)
    )


    tree.focus()
    loadMetricsTree()
    fetchMetricData selected_metrics, time_from

