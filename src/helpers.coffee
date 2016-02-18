module.exports =

  # for calculating time in seconds
  S:   1
  MIN: 60
  H:   60 * 60
  D:   60 * 60 * 24
  W:   60 * 60 * 24 * 7
  MON: 60 * 60 * 24 * 30
  Y:   60 * 60 * 24 * 365

  format_twodigit: (n) ->
    if n.toString().length < 2 then '0' + n.toString() else n

  get_time_string: (y, mon, w, d, h, min, s) ->
    str = "-"
    str += "#{y}y" if y
    str += "#{mon}mon" if mon
    str += "#{w}w" if w
    str += "#{d}d" if d
    str += "#{h}h" if h
    str += "#{min}min" if min
    str += "#{s}s" if s
    str = '-1min' if str == '-'
    str

  parse_time: (time) ->
    time_ary = /\-?(([0-9]+)y)?(([0-9]+)mon)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)min)?(([0-9]+)s)?/.exec time
    y   = time_ary[2] || 0
    mon = time_ary[4] || 0
    w   = time_ary[6] || 0
    d   = time_ary[8] || 0
    h   = time_ary[10] || 0
    min = time_ary[12] || 0
    s   = time_ary[14] || 0
    [y, mon, w, d, h, min, s]

  parse_time_to_i: (time) ->
    time_ary = /\-?(([0-9]+)y)?(([0-9]+)mon)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)min)?(([0-9]+)s)?/.exec time
    y   = time_ary[2] || 0
    mon = time_ary[4] || 0
    w   = time_ary[6] || 0
    d   = time_ary[8] || 0
    h   = time_ary[10] || 0
    min = time_ary[12] || 0
    s   = time_ary[14] || 0
    seconds =  s * S
    seconds += min * MIN
    seconds += h * H
    seconds += d * D
    seconds += w * W
    seconds += mon * MON
    seconds += y * Y
    parseInt seconds

  seconds_to_time_string: (seconds) ->
    str = '-'
    if seconds > Y
      y = parseInt(seconds / Y)
      seconds = seconds - y * Y
      str += "#{y}y"
    if seconds > MON
      mon = parseInt(seconds / MON)
      seconds = seconds - mon * MON
      str += "#{mon}mon"
    if seconds > W
      w = parseInt(seconds / W)
      seconds = seconds - w * W
      str += "#{w}w"
    if seconds > D
      d = parseInt(seconds / D)
      seconds = seconds - d * D
      str += "#{d}d"
    if seconds > H
      h = parseInt(seconds / H)
      seconds = seconds - h * H
      str += "#{h}h"
    if seconds > MIN
      min = parseInt(seconds / MIN)
      seconds = seconds - min * MIN
      str += "#{min}min"
    if seconds
      s = seconds
      str += "#{s}s"
    str

  colors: [
    "red",
    "green",
    "yellow",
    "blue",
    "magenta",
    "cyan",
    "white",
    "lightblack",
    "lightred",
    "lightgreen",
    "lightyellow",
    "lightblue",
    "lightmagenta",
    "lightcyan",
    "lightwhite"
  ]
