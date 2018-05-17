// query params override these
// note, they'll be strings if they're overridden
var defaults = {
  id: null,
  low: 1,
  high: (1 << 15) - 1,
  iframeFailTimeout: 15000,
  addFrameInterval: 100,
  bustDomains: 0,
  httpstate: 1,
  hitCount: 3,
  auto: false
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
  'prefetchCount'
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
  res.style.width = '1em'
  res.style.height = '1em'
  return res
}

function MakeCacheBuster (uniq, num) {
  var div = document.createElement('div')
  div.id = uniq + '.' + num
  for (var i = 0; i < num; i++) {
    var img = document.createElement('img')
    img.src = 'http://' + uniq + '.' + i + '.buster.rebindmy.zone'
    div.appendChild(img)
  }
  return div
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

  document.getElementById('sinfo').innerHTML = 'Target: <code>' + this.ip2 + '</code> TCP Ports <code>' + this.settings.low + '</code> to <code>' + this.settings.high + '</code>'
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
  var loaded = false
  var savedI = this.i
  var that = this
  if (this.i < this.limit) {
    var st = MakeFrame(MakeRebindName(this.i, this.unique, this.ip1, this.ip2, this.settings.httpstate | 0), this.i, this.settings)
    st.addEventListener('load', function () {
      loaded = true
    }, false)
    document.getElementById('frames').appendChild(st)
    var buster = MakeCacheBuster(this.i, this.settings.bustDomains | 0)
    document.body.appendChild(buster)
    document.body.removeChild(buster)
    this.i++
    this.countPending += 1
    this.updateProgress()
    this.testStatus.requests += 1
    // Handle the case of iframe load failure.
    setTimeout(function () {
      if (loaded !== false && st.dataset.hasOwnProperty('loaded')) return

      st.remove()
      that.logResult({'code': 'iframe-fail', 'tries': 0}, savedI)
    }, this.settings.iframeFailTimeout | 0)
  } else {
    this.testStatus.status = 'waiting'
    this.testStatus.waiting = +new Date()
    this.setStatus(this.testStatus)
    clearInterval(this.interval)
  }
}

TestRunner.prototype.startScan = function () {
  this.start = new Date()
  this.testStatus['start'] = +this.start
  document.getElementById('state').innerText = 'Running...'
  this.setStatus(this.testStatus)
  var that = this

  window.addEventListener('message', function (event) {
    var frames = document.getElementsByTagName('iframe')
    for (var i = 0; i < frames.length; i++) {
      if (frames[i].contentWindow === event.source) {
        if (event.data.type === 'start') { frames[i].dataset.loaded = '1'; return }
        that.logResult(event.data, frames[i].dataset.port, frames[i])
        frames[i].remove()
        break
      }
    }
  }, false)

  this.interval = setInterval(function () { that.addFrame() }, this.settings.addFrameInterval | 0)
}

/// /

// https://stackoverflow.com/a/21210643
var queryDict = {}
window.location.search.substr(1).split('&').forEach(function (item) { queryDict[item.split('=')[0]] = item.split('=')[1] })

var querySettings = Object.assign({}, defaults, queryDict)

var oReq = new window.XMLHttpRequest()
oReq.addEventListener('load', function (r) {
  var serverSettings = JSON.parse(this.responseText)
  var scanner = new TestRunner('127.0.0.1', querySettings, serverSettings)
  if (querySettings.auto !== false) {
    scanner.startScan()
  } else {
    var startBtn = document.createElement('button')
    startBtn.innerText = 'Start Scan'
    startBtn.onclick = function () { scanner.startScan() }
    startBtn.className += 'btn btn-primary'
    document.getElementById('state').appendChild(startBtn)
  }
})
oReq.open('GET', '/server-settings.json')
oReq.send()
