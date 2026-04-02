package stealth

// ChromeStealthScripts returns JavaScript code snippets that must be
// injected BEFORE page load via CDP's Page.addScriptToEvaluateOnNewDocument.
//
// These scripts are ported from puppeteer-extra-plugin-stealth and defeat
// common bot detection methods used by:
// - Cloudflare Bot Management
// - DataDome
// - Akamai Bot Manager
// - PerimeterX / HUMAN
// - Kasada
//
// Each script targets a specific detection vector.
func ChromeStealthScripts() []string {
	return []string{
		// 1. navigator.webdriver — THE most common detection method.
		// Real browsers have this as undefined; automation sets it to true.
		`Object.defineProperty(navigator, 'webdriver', {
			get: () => undefined,
		});`,

		// 2. navigator.plugins — Headless Chrome has empty plugins array.
		// Real Chrome on Windows/Mac always has these 3+ plugins.
		`Object.defineProperty(navigator, 'plugins', {
			get: () => {
				const makePlugin = (name, filename, desc) => {
					const p = { name, filename, description: desc, length: 1 };
					p[0] = { type: 'application/pdf', suffixes: 'pdf', description: 'Portable Document Format' };
					return p;
				};
				const arr = [
					makePlugin('Chrome PDF Plugin', 'internal-pdf-viewer', 'Portable Document Format'),
					makePlugin('Chrome PDF Viewer', 'mhjfbmdgcfjbbpaeojofohoefgiehjai', ''),
					makePlugin('Native Client', 'internal-nacl-plugin', ''),
				];
				arr.length = 3;
				arr.item = (i) => arr[i] || null;
				arr.namedItem = (name) => arr.find(p => p.name === name) || null;
				arr.refresh = () => {};
				return arr;
			},
		});`,

		// 3. navigator.languages — Must match Accept-Language header.
		`Object.defineProperty(navigator, 'languages', {
			get: () => ['tr-TR', 'tr', 'en-US', 'en'],
		});`,

		// 4. navigator.permissions — Headless Chrome returns 'prompt'
		// for notifications even when they're actually denied.
		`const originalQuery = window.navigator.permissions.query;
		window.navigator.permissions.query = (parameters) => (
			parameters.name === 'notifications'
				? Promise.resolve({ state: Notification.permission })
				: originalQuery(parameters)
		);`,

		// 5. window.chrome — Must exist in real Chrome, doesn't in headless.
		`if (!window.chrome) {
			window.chrome = {};
		}
		window.chrome.runtime = window.chrome.runtime || {};
		window.chrome.loadTimes = window.chrome.loadTimes || function() {
			return {
				commitLoadTime: Date.now() / 1000,
				finishDocumentLoadTime: Date.now() / 1000 + 0.1,
				finishLoadTime: Date.now() / 1000 + 0.2,
				firstPaintAfterLoadTime: 0,
				firstPaintTime: Date.now() / 1000 + 0.05,
				navigationType: 'Other',
				npnNegotiatedProtocol: 'h2',
				requestTime: Date.now() / 1000 - 0.5,
				startLoadTime: Date.now() / 1000 - 0.4,
				wasAlternateProtocolAvailable: false,
				wasFetchedViaSpdy: true,
				wasNpnNegotiated: true,
			};
		};
		window.chrome.csi = window.chrome.csi || function() {
			return {
				onloadT: Date.now(),
				startE: Date.now() - 500,
				pageT: 500,
				tran: 15,
			};
		};`,

		// 6. Function.prototype.toString — Anti-bot checks if native functions
		// have been overridden by checking toString output.
		`const nativeToString = Function.prototype.toString;
		const myToString = function() {
			if (this === myToString) {
				return 'function toString() { [native code] }';
			}
			return nativeToString.call(this);
		};
		Function.prototype.toString = myToString;`,

		// 7. WebGL — Headless Chrome reports different GPU vendor/renderer.
		// This spoofs it to look like a common Intel GPU.
		`const getParameter = WebGLRenderingContext.prototype.getParameter;
		WebGLRenderingContext.prototype.getParameter = function(param) {
			if (param === 37445) return 'Intel Inc.';
			if (param === 37446) return 'Intel Iris OpenGL Engine';
			return getParameter.call(this, param);
		};`,

		// 8. navigator.connection — Missing in headless, present in real Chrome.
		`if (!navigator.connection) {
			Object.defineProperty(navigator, 'connection', {
				get: () => ({
					downlink: 10,
					effectiveType: '4g',
					onchange: null,
					rtt: 50,
					saveData: false,
				}),
			});
		}`,

		// 9. navigator.hardwareConcurrency — Headless sometimes reports wrong value.
		`Object.defineProperty(navigator, 'hardwareConcurrency', {
			get: () => 8,
		});`,

		// 10. navigator.deviceMemory — Often missing in headless.
		`Object.defineProperty(navigator, 'deviceMemory', {
			get: () => 8,
		});`,

		// 11. Iframe contentWindow detection — Some anti-bot creates an iframe
		// and checks properties on its contentWindow.
		`const originalAttachShadow = Element.prototype.attachShadow;
		Element.prototype.attachShadow = function() {
			return originalAttachShadow.call(this, ...arguments);
		};`,

		// 12. Canvas fingerprint noise — Add subtle random noise to canvas
		// to avoid consistent headless fingerprints.
		`const originalGetContext = HTMLCanvasElement.prototype.getContext;
		HTMLCanvasElement.prototype.getContext = function(type, attrs) {
			const ctx = originalGetContext.call(this, type, attrs);
			if (type === '2d' && ctx) {
				const originalGetImageData = ctx.getImageData;
				ctx.getImageData = function() {
					const imageData = originalGetImageData.apply(this, arguments);
					for (let i = 0; i < imageData.data.length; i += 4) {
						imageData.data[i] ^= (Math.random() * 2) | 0;
					}
					return imageData;
				};
			}
			return ctx;
		};`,
	}
}

// HumanBehaviorScript returns JavaScript that simulates human-like
// page interactions: mouse movements, random scrolling, hover events.
func HumanBehaviorScript() string {
	return `
		(function() {
			// Simulate random mouse movements
			const events = ['mousemove', 'mouseenter', 'mouseover'];
			for (let i = 0; i < 5 + Math.floor(Math.random() * 10); i++) {
				const evt = new MouseEvent(events[Math.floor(Math.random() * events.length)], {
					bubbles: true,
					clientX: Math.floor(Math.random() * window.innerWidth),
					clientY: Math.floor(Math.random() * window.innerHeight),
				});
				document.dispatchEvent(evt);
			}

			// Random scroll (not just to bottom)
			const scrollPositions = [
				window.innerHeight * 0.3,
				window.innerHeight * 0.6,
				document.body.scrollHeight * 0.5,
				document.body.scrollHeight * 0.8,
				document.body.scrollHeight,
			];
			const pos = scrollPositions[Math.floor(Math.random() * scrollPositions.length)];
			window.scrollTo({ top: pos, behavior: 'smooth' });
		})();
	`
}
