var reCAPTCHA_ok = false;

var form = document.getElementById('form');
var passInput = document.getElementById('password');
var pass2Input = document.getElementById('password2');
var submitButton = document.getElementById('submitButton');

passInput.addEventListener('keydown', function(e) {
	if (e.key === 'Enter') {
		e.preventDefault();
		if (passInput.value.trim() !== '') {
			pass2Input.focus();
		}
	}
});

pass2Input.addEventListener('keydown', function(e) {
	if (e.key === 'Enter') {
		e.preventDefault();
		if (pass2Input.value.trim() !== '') {
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
	passInput.focus();
});
