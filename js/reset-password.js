var reCAPTCHA_ok = false;

var form = document.getElementById('form');
var newPassInput = document.getElementById('newpass');
var newPass2Input = document.getElementById('newpass2');
var submitButton = document.getElementById('submitButton');

newPassInput.addEventListener('keydown', function(e) {
	if (e.keyCode === 13) {
		e.preventDefault();
		if (newPassInput.value.trim() !== '') {
			newPass2Input.focus();
		}
	}
});

newPass2Input.addEventListener('keydown', function(e) {
	if (e.keyCode === 13) {
		e.preventDefault();
		if (newPass2Input.value.trim() !== '') {
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
	newPassInput.focus();
});
