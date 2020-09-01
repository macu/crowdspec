<template>
<el-dialog
	:title="'Your username is ' + username"
	:visible.sync="showing"
	:width="$store.getters.dialogTinyWidth"
	:close-on-click-modal="false"
	@closed="clearMode()"
	class="edit-profile-modal">

	<el-form v-if="changePasswordMode"
		ref="changePasswordForm"
		:model="changePasswordForm"
		:rules="changePasswordRules"
		v-loading="sending"
		label-position="top">
		<el-form-item label="Current password" prop="oldPass">
			<el-input
				ref="oldPass"
				type="password"
				name="old_password"
				v-model="changePasswordForm.oldPass"
				autocomplete="off"
				@keyup.enter.native="handleChangePasswordReturn($refs.oldPass)"/>
		</el-form-item>
		<el-form-item label="New password" prop="newPass">
			<el-input
				ref="newPass"
				type="password"
				name="new_password"
				v-model="changePasswordForm.newPass"
				autocomplete="off"
				@keyup.enter.native="handleChangePasswordReturn($refs.newPass)"/>
		</el-form-item>
		<el-form-item label="Confirm new password" prop="newPass2">
			<el-input
				ref="newPass2"
				type="password"
				name="verify_password"
				v-model="changePasswordForm.newPass2"
				autocomplete="off"
				@keyup.enter.native="handleChangePasswordReturn($refs.newPass2)"/>
		</el-form-item>
		<el-form-item>
			<el-alert type="info" :closable="false">
				<p>Your password is sent through HTTPS to Google Cloud Platform in Montr&eacute;al and encrypted server-side using <a href="https://pkg.go.dev/golang.org/x/crypto/bcrypt?tab=doc#GenerateFromPassword" target="_blank">bcrypt</a>.</p>
				<p>Please use a good password so your account is less likely to be hijacked and used to cause trouble.</p>
				<p>
					For your security, you should not use the same passwords that you use for important services such as email and online banking on other websites.
					If one website gets hacked, you can lose access to every other site where you use the same password.
					The best setup is to use a password manager with two-factor authentication and distinct long random passwords for every important service.
				</p>
			</el-alert>
		</el-form-item>
		<el-form-item>
			<el-button type="primary" @click="submitChangePassword()">
				Update
			</el-button>
			<el-button @click="clearMode()">
				Cancel
			</el-button>
		</el-form-item>
	</el-form>
	<el-button v-else
		@click="enterChangePasswordMode()"
		class="change-password">
		Change password
	</el-button>

	<span slot="footer" class="dialog-footer">
		<el-button @click="showing = false">Close</el-button>
	</span>

</el-dialog>
</template>

<script>
import $ from 'jquery';
import {alertError} from '../utils.js';

const MODE_CHANGE_PASSWORD = 1;

const SUCCESS_TIMEOUT = 1200;
const ERROR_TIMEOUT = 4000;

export default {
	data() {
		return {
			showing: false,
			mode: null,
			changePasswordForm: {
				oldPass: '',
				newPass: '',
				newPass2: '',
			},
			sending: false,
		};
	},
	computed: {
		userID() {
			return this.$store.getters.userID;
		},
		username() {
			return this.$store.getters.username;
		},
		changePasswordMode() {
			return this.mode === MODE_CHANGE_PASSWORD;
		},
		changePasswordRules() {
			return {
				oldPass: [{
					validator: (rule, value, callback) => {
						if (!value.trim()) {
							callback(new Error('Enter your old password'));
						} else {
							callback();
						}
					},
					trigger: 'blur',
				}],
				newPass: [{
					validator: (rule, value, callback) => {
						if (value.trim().length < 5) {
							callback(new Error('Password minimum length is 5 digits'));
						} else {
							if (this.changePasswordForm.newPass2.trim()) {
								this.$refs.changePasswordForm.validateField('newPass2');
							}
							callback();
						}
					},
					trigger: 'blur',
				}],
				newPass2: [{
					validator: (rule, value, callback) => {
						if (!value.trim()) {
							callback(new Error('Repeat new password'));
						} else if (value !== this.changePasswordForm.newPass) {
							callback(new Error('Passwords do not match'));
						} else {
							callback();
						}
					},
					trigger: 'blur',
				}],
			};
		},
	},
	methods: {
		show() {
			this.showing = true;
		},
		enterChangePasswordMode() {
			this.mode = MODE_CHANGE_PASSWORD;
			this.selectOldPassword();
		},
		handleChangePasswordReturn() {
			if (!this.changePasswordForm.oldPass.trim()) {
				$('input', this.$refs.oldPass.$el).focus().select();
			} else if (!this.changePasswordForm.newPass.trim()) {
				$('input', this.$refs.newPass.$el).focus().select();
			} else if (!this.changePasswordForm.newPass2.trim() ||
				this.changePasswordForm.newPass !== this.changePasswordForm.newPass2) {
				$('input', this.$refs.newPass2.$el).focus().select();
			} else {
				this.submitChangePassword();
			}
		},
		submitChangePassword() {
			this.$refs.changePasswordForm.validate(valid => {
				if (valid) {
					this.sending = true;
					$.post('/ajax/user/change-password', {
						old: this.changePasswordForm.oldPass,
						new: this.changePasswordForm.newPass,
						new2: this.changePasswordForm.newPass2,
					}).then(() => {
						this.$message({
							message: 'Password updated successfully',
							type: 'success',
							showClose: true,
							duration: SUCCESS_TIMEOUT,
						});
						this.clearMode();
					}).fail(error => {
						if (error && error.readyState) {
							switch (error.status) {
								case 400: // Bad request
									this.$message({
										message: 'New password rejected',
										type: 'error',
										showClose: true,
										duration: ERROR_TIMEOUT,
									});
									break;
								case 403: // Forbidden
									this.$message({
										message: 'Current password incorrect',
										type: 'error',
										showClose: true,
										duration: ERROR_TIMEOUT,
									});
									this.selectOldPassword();
									break;
								default:
									alertError(error);
							}
						} else {
							alertError(error);
						}
					}).always(() => {
						this.sending = false;
					});
				}
			});
		},
		selectOldPassword() {
			this.$nextTick(() => {
				if (this.$refs.oldPass) {
					$('input', this.$refs.oldPass.$el).focus().select();
				}
			});
		},
		clearMode() {
			this.mode = null;
			this.changePasswordForm = {
				oldPass: '',
				newPass: '',
				newPass2: '',
			};
		},
	},
};
</script>

<style lang="scss">
.edit-profile-modal {
	.el-form {
		.el-form-item {
			margin-bottom: 30px;
			.el-form-item__label {
				font-size: 1em;
				line-height: 1.2em;
			}
			.el-alert {
				>.el-alert__content {
					line-height: 1.7em;
					>.el-alert__description {
						>p {
							margin-top: 0;
							margin-bottom: 0;
						}
						>p + p {
							margin-top: 5px;
						}
					}
				}
			}
		}
	}
}
</style>
