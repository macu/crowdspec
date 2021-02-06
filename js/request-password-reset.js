var reCAPTCHA_ok = false;

var form = document.getElementById('form');
var emailInput = document.getElementById('email');
var submitButton = document.getElementById('submitButton');

emailInput.addEventListener('keydown', function(e) {
	if (e.key === 'Enter') {
		e.preventDefault();
		if (emailInput.value.trim() !== '') {
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
	emailInput.focus();
});
