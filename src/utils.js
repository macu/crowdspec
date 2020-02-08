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
