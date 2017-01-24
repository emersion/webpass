const CACHE_VERSION = 1
const CACHE_NAME = 'static-v'+CACHE_VERSION

self.addEventListener('activate', event => {
	console.log('ServiceWorker activated')

	event.waitUntil(
		caches.keys().then(names => {
			return Promise.all(names.map(name => {
				if (name !== CACHE_NAME) {
					return caches.delete(name)
				}
			}))
		})
	)
})

self.addEventListener('install', event => {
	console.log('ServiceWorker installed')

	return caches.open(CACHE_NAME).then(cache => {
		return cache.addAll(['/'])
	})
})

self.addEventListener('fetch', event => {
	const req = event.request
	console.log(req.url)
	if (req.method !== 'GET') {
		return event.respondWith(fetch(req))
	}

	event.respondWith(
		fetch(req.clone())
		.catch(err => {
			return caches.open(CACHE_NAME).then(cache => {
				return cache.match(req)
			})
			.then(cached => {
				if (cached) {
					console.log('Serving from cache because of a network error for '+req.url)
					return cached
				}
				console.log('Response not in cache for '+req.url)
				throw err
			})
		})
		.then(res => {
			if (res.status >= 400 || !res.headers.has('content-type')) {
				return res
			}

			const mediaType = res.headers.get('content-type')
			if (mediaType !== 'text/html' && mediaType !== 'text/css' && mediaType !== 'application/x-javascript') {
				return res
			}

			return caches.open(CACHE_NAME).then(cache => {
				return cache.match(req)
				.then(cached => {
					if (!cached) {
						console.log('Caching response for '+req.url)
						cache.put(req, res.clone())
						return res
					}

					console.log('Comparing online and cached response for '+req.url)
					return Promise.all([
						res.arrayBuffer(),
						cached.clone().arrayBuffer(),
					]).then(bufs => {
						if (bufs[0].byteLength !== bufs[1].byteLength) {
							return false
						}

						const resDataView = new DataView(bufs[0])
						const cachedDataView = new DataView(bufs[1])
						for (let i = 0; i < resDataView.byteLength; i++) {
							if (resDataView.getInt8(i) !== cachedDataView.getInt8(i)) {
								return false
							}
						}
						return true
					})
					.then(equals => {
						if (!equals) {
							console.log('Response changed!')
						} else {
							console.log('Response did not change, serving from cache for '+req.url)
						}
						return cached
					})
				})
			})
		})
	)
})
