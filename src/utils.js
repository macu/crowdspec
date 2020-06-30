import $ from 'jquery';
import Vue from 'vue';

export function alertError(error) {
	console.error(error);
	if (error) {
		let message = null;
		if (error.readyState === 0) {
			message = 'Could not connect to server';
		} else if (error.readyState && error.status) {
			message = 'Request failed with error code ' + error.status;
		} else if (error.message) {
			message = error.message;
		} else if (typeof error === 'string') {
			message = error;
		}
		if (message) {
			Vue.prototype.$alert(message, 'Error', {
				confirmButtonText: 'Ok',
				type: 'error',
			});
			return;
		}
	}
	Vue.prototype.$alert('An error occurred', {
		confirmButtonText: 'Ok',
		type: 'error',
	});
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
		console.log('requesting');
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
