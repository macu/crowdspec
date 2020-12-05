import $ from 'jquery';
import Vue from 'vue';

const VERSION_STAMP_RESPONSE = /^VersionStamp: (.+)$/m;

export function alertError(error) {
	console.error(error);
	let message = null;
	if (error) {
		if (error.readyState === 0) {
			message = 'Could not connect to server.';
		} else if (error.readyState && error.status) {
			message = 'Request failed with HTTP error code ' + error.status + '.';
		} else if (error.message) {
			message = error.message;
		} else if (typeof error === 'string') {
			message = error;
		}
	}
	if (!message) {
		message = 'An error occurred.';
	}
	if (error && error.responseText &&
			VERSION_STAMP_RESPONSE.test(error.responseText)) {
		let match = VERSION_STAMP_RESPONSE.exec(error.responseText);
		if (match[1] !== window.appVersion) {
			message += ' (A new version is available. Reload this page to use the latest client code.)'
		}
	}
	Vue.prototype.$alert(message, 'Error', {
		confirmButtonText: 'Ok',
		type: 'error',
	});
}

export function defaultUserSettings() {
	return {
		blockEditing: {
			deleteButton: 'all',
		},
	};
}

// Use in scenarios where comparing numeric IDs of mixed type (string / int).
// E.g., objects from the server use numeric IDs, but $route params are strings.
export function idsEq(id1, id2) {
	return parseInt(id1, 10) === parseInt(id2, 10);
}

// Returns a function that invokes the given callback after the specified delay,
// unless the returned function is called again during the delay and then the delay is extended.
export function debounce(callback, timeoutMs = 500) {
	var timeout;
	function invoker() {
		timeout = null;
		callback.apply(this, arguments);
	}
	return function() {
		if (timeout) {
			clearTimeout(timeout);
		}
		return timeout = setTimeout(
			invoker.bind(this, ...arguments),
			timeoutMs
		);
	};
}

export function setWindowSubtitle(subtitle = null) {
	window.title = 'CrowdSpec' + (subtitle ? ' | ' : '') + subtitle;
}

// Call when dragging starts, returns handler.
// Call handler.stop() when dragging stops.
export function startAutoscroll() {
	const GUTTER_SIZE = 70; // distance from edge of viewport where scrolling starts
	const SCALE_RANGE = 8; // higher value gives potential for faster scrolling

	const $window = $(window);

	let requestId = null;
	let clientY = null; // cursor position within viewport

	function handleMouseMove(e) {
		clientY = e.clientY;
	}

	$window.on('mousemove', handleMouseMove);

	function handleScroll() {
		if (clientY !== null) {
			let viewportHeight = $window.height(), delta = 0;
			if (clientY < GUTTER_SIZE) { // Scroll up
				let factor = (GUTTER_SIZE - clientY) / GUTTER_SIZE;
				delta = -((factor * SCALE_RANGE) + 1);
			} else if (clientY > (viewportHeight - GUTTER_SIZE)) { // Scroll down
				let factor = (clientY - (viewportHeight - GUTTER_SIZE)) / GUTTER_SIZE;
				delta = (factor * SCALE_RANGE) + 1;
			}
			if (delta !== 0) {
				$window.scrollTop($window.scrollTop() + delta);
			}
		}
		requestId = window.requestAnimationFrame(handleScroll);
	}

	requestId = window.requestAnimationFrame(handleScroll);

	return {
		stop: function() {
			$window.off('mousemove', handleMouseMove);
			if (requestId) {
				window.cancelAnimationFrame(requestId);
				requestId = null;
			}
		},
	};
}

export function isValidURL(url) {
	if (window.URL) {
		try {
			new URL(url);
			return true;
		} catch {
			return false;
		}
	}
	// No URL in IE
	// TODO fallback on regex
	return !!(url && url.trim());
}

// Recognized video URL formats:
// http://www.youtube.com/watch?v=My2FRPA3Gf8
// http://www.youtube.com/embed/My2FRPA3Gf8
// http://youtu.be/My2FRPA3Gf8
// https://youtube.googleapis.com/v/My2FRPA3Gf8
// http://vimeo.com/25451551
// http://player.vimeo.com/video/25451551
export const VID_URL_REGEX = /^https?:\/\/(?:www\.|m\.)?(?:youtube\.com\/watch\?v=|youtube\.com\/embed\/|youtu\.be\/|youtube\.googleapis\.com\/v\/|vimeo\.com\/|player\.vimeo\.com\/video\/)([a-zA-Z0-9_-]+)/;

export function isVideoURL(url) {
	return VID_URL_REGEX.test(url);
}

export function extractVid(url) {
	let match = VID_URL_REGEX.exec(url);
	if (!match) {
		return null;
	}
	if (url.indexOf('youtube.com') || url.indexOf('youtu.be')) {
		return {type: 'youtube', id: match[1]};
	} else if (url.indexOf('vimeo.com')) {
		return {type: 'vimeo', id: match[1]};
	}
	return null;
}
