var reCAPTCHA_ok = false;

var form = document.getElementById('form');
var usernameInput = document.getElementById('username');
var passwordInput = document.getElementById('password');
var submitButton = document.getElementById('submitButton');

usernameInput.addEventListener('keydown', function(e) {
	if (e.key === 'Enter') {
		e.preventDefault();
		if (usernameInput.value.trim() !== '') {
			passwordInput.focus();
		}
	}
});

passwordInput.addEventListener('keydown', function(e) {
	if (e.key === 'Enter') {
		e.preventDefault();
		if (passwordInput.value.trim() !== '') {
			if (reCAPTCHA_ok || !window.verifyRequired) {
				form.submit();
			}
		}
	}
});

submitButton.addEventListener('click', function(e) {
	e.preventDefault();
	if (reCAPTCHA_ok || !window.verifyRequired) {
		form.submit();
	}
});

function recaptchaSuccess() {
	reCAPTCHA_ok = true;
	submitButton.disabled = false;
}

function recaptchaExpired() {
	reCAPTCHA_ok = false;
	if (window.verifyRequired) {
		submitButton.disabled = true;
	}
}

function recaptchaError() {
	reCAPTCHA_ok = false;
	if (window.verifyRequired) {
		submitButton.disabled = true;
	}
}

document.addEventListener('DOMContentLoaded', function() {
	usernameInput.focus();
});
