var reCAPTCHA_ok = false;

var form = document.getElementById('form');
var usernameInput = document.getElementById('username');
var passwordInput = document.getElementById('password');
var submitButton = document.getElementById('submitButton');

usernameInput.addEventListener('keydown', function(e) {
	if (e.keyCode === 13) {
		e.preventDefault();
		if (usernameInput.value.trim() !== '') {
			passwordInput.focus();
		}
	}
});

passwordInput.addEventListener('keydown', function(e) {
	if (e.keyCode === 13) {
		e.preventDefault();
		if (passwordInput.value.trim() !== '') {
			if (reCAPTCHA_ok) {
				form.submit();
			}
		}
	}
});

submitButton.addEventListener('click', function(e) {
	e.preventDefault();
	if (reCAPTCHA_ok) {
		form.submit();
	}
});

function recaptchaSuccess() {
	reCAPTCHA_ok = true;
	submitButton.disabled = false;
}

function recaptchaExpired() {
	reCAPTCHA_ok = false;
	submitButton.disabled = true;
}

function recaptchaError() {
	reCAPTCHA_ok = false;
	submitButton.disabled = true;
}

document.addEventListener('DOMContentLoaded', function() {
	usernameInput.focus();
});
