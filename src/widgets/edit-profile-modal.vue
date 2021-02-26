<template>
<el-dialog
	:visible.sync="showing"
	:width="$store.getters.dialogTinyWidth"
	:close-on-click-modal="mode === null"
	@closed="clearMode()"
	class="edit-profile-modal">

	<span slot="title">
		Account options for
		<username :username="username" :highlight="highlight"/>
	</span>

	<el-form v-if="updateSettingsMode"
		ref="changePasswordForm"
		:model="settingsForm"
		v-loading="waiting"
		label-position="top">
		<el-form-item>
			<strong slot="label" class="section-heading">Profile</strong>
			<div class="flex-input">
				<el-color-picker
					v-model="settingsForm.userProfile.highlightUsername"
					:predefine="predefinedUsernameHighlights"
					color-format="rgb"
					/>
				<el-button
					v-if="settingsForm.userProfile.highlightUsername"
					@click="settingsForm.userProfile.highlightUsername = null"
					size="mini"
					icon="el-icon-close"
					/>
				<username username="Username colour" :highlight="settingsForm.userProfile.highlightUsername"/>
			</div>
		</el-form-item>
		<el-form-item>
			<strong slot="label" class="section-heading">Block editing</strong>
			<el-select v-model="settingsForm.blockEditing.deleteButton">
				<el-option label="Show delete button only in edit block modal" value="modal"/>
				<el-option label="Show delete button on newly added blocks" value="recent"/>
				<el-option label="Show delete button on all blocks" value="all"/>
			</el-select>
		</el-form-item>
		<el-form-item>
			<strong slot="label" class="section-heading">Community</strong>
			<el-checkbox v-model="settingsForm.community.unreadOnly">
				Show only unread comments by default
			</el-checkbox>
		</el-form-item>
		<el-form-item>
			<el-button type="primary" @click="submitSettings()">
				Update
			</el-button>
			<el-button @click="clearMode()">
				Cancel
			</el-button>
		</el-form-item>
	</el-form>

	<el-form v-else-if="changePasswordMode"
		ref="changePasswordForm"
		:model="changePasswordForm"
		:rules="changePasswordRules"
		v-loading="waiting"
		label-position="top">
		<el-form-item label="Current password" prop="oldPass">
			<el-input
				ref="oldPass"
				type="password"
				name="old_password"
				v-model="changePasswordForm.oldPass"
				autocomplete="current-password"
				@keyup.enter.native="handleChangePasswordReturn($refs.oldPass)"/>
		</el-form-item>
		<el-form-item label="New password" prop="newPass">
			<el-input
				ref="newPass"
				type="password"
				name="new_password"
				v-model="changePasswordForm.newPass"
				autocomplete="new-password"
				@keyup.enter.native="handleChangePasswordReturn($refs.newPass)"/>
		</el-form-item>
		<el-form-item label="Confirm new password" prop="newPass2">
			<el-input
				ref="newPass2"
				type="password"
				name="verify_password"
				v-model="changePasswordForm.newPass2"
				autocomplete="new-password"
				@keyup.enter.native="handleChangePasswordReturn($refs.newPass2)"/>
		</el-form-item>
		<el-form-item>
			<el-alert type="info" :closable="false">
				<p>Your password is sent through HTTPS to Google Cloud Platform in Montr&eacute;al and encrypted server-side using <a href="https://pkg.go.dev/golang.org/x/crypto/bcrypt?tab=doc#GenerateFromPassword" target="_blank">bcrypt</a>.</p>
				<p>Please use a good password so your account is less likely to be hijacked and used to cause trouble.</p>
				<p>
					For your security, you should not use the same passwords that you use for important services such as email and online banking on other websites.
					If one website gets hacked, you can lose access to every other site where you use the same password.
					Is is wise to use a password manager with two-factor authentication and distinct long random passwords for every important service.
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

	<div v-else class="options">
		<p>Your email address on file is <span class="email">{{$store.getters.emailAddress}}</span>.</p>
		<el-button @click="enterUpdateSettingsMode()">
			Update settings
		</el-button>
		<el-button @click="enterChangePasswordMode()">
			Change password
		</el-button>
	</div>

	<span slot="footer" class="dialog-footer">
		<el-button @click="showing = false">Close</el-button>
	</span>

</el-dialog>
</template>

<script>
import $ from 'jquery';
import Username from './username.vue';
import {alertError, defaultUserSettings} from '../utils.js';

const MODE_UPDATE_SETTINGS = 1;
const MODE_CHANGE_PASSWORD = 2;

const SUCCESS_MESSAGE_TIMEOUT = 1200;
const ERROR_MESSAGE_TIMEOUT = 4000;

export default {
	components: {
		Username,
	},
	data() {
		return {
			showing: false,
			mode: null,
			settingsForm: defaultUserSettings(),
			changePasswordForm: {
				oldPass: '',
				newPass: '',
				newPass2: '',
			},
			waiting: false,
		};
	},
	computed: {
		username() {
			return this.$store.getters.username;
		},
		highlight() {
			return this.$store.getters.userSettings.userProfile.highlightUsername;
		},
		updateSettingsMode() {
			return this.mode === MODE_UPDATE_SETTINGS;
		},
		changePasswordMode() {
			return this.mode === MODE_CHANGE_PASSWORD;
		},
		predefinedUsernameHighlights() {
			return [
				'rgb(50, 205, 50)', // limegreen
				'rgb(255, 192, 203)', // pink
				'rgb(240, 255, 255)', // azure
				'rgb(255, 255, 0)', // yellow
				'rgb(255, 250, 250)', // snow
			];
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
						if (value.trim().length < window.const.passwordMinLength) {
							callback(new Error('Password minimum length is ' +
								window.const.passwordMinLength + ' digits'));
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
		enterUpdateSettingsMode() {
			this.mode = MODE_UPDATE_SETTINGS;
			this.waiting = true;
			$.get('/ajax/settings').then(settings => {
				this.settingsForm = $.extend(true, defaultUserSettings(), settings);
				this.waiting = false;
			}).fail(jqXHR => {
				this.waiting = false;
				this.clearMode();
				alertError(jqXHR);
			});
		},
		enterChangePasswordMode() {
			this.mode = MODE_CHANGE_PASSWORD;
			this.selectOldPassword();
		},
		submitSettings() {
			this.waiting = true;
			$.post('/ajax/user/save-settings', {
				settings: JSON.stringify(this.settingsForm),
			}).then(settings => {
				this.waiting = false;
				this.$message({
					message: 'Settings updated',
					type: 'success',
					showClose: true,
					duration: SUCCESS_MESSAGE_TIMEOUT,
				});
				this.clearMode();
				this.$store.commit('setUserSettings', settings);
			}).fail(jqXHR => {
				this.waiting = false;
				alertError(jqXHR);
			});
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
					this.waiting = true;
					$.post('/ajax/user/change-password', {
						old: this.changePasswordForm.oldPass,
						new: this.changePasswordForm.newPass,
						new2: this.changePasswordForm.newPass2,
					}).then(() => {
						this.$message({
							message: 'Password updated successfully',
							type: 'success',
							showClose: true,
							duration: SUCCESS_MESSAGE_TIMEOUT,
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
										duration: ERROR_MESSAGE_TIMEOUT,
									});
									break;
								case 403: // Forbidden
									this.$message({
										message: 'Current password incorrect',
										type: 'error',
										showClose: true,
										duration: ERROR_MESSAGE_TIMEOUT,
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
						this.waiting = false;
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
@import '../_styles/_app.scss';

.edit-profile-modal {

	.flex-input {
		display: flex;
		flex-direction: row;
		align-items: center;

		>:not(:first-child) {
			margin-left: 10px;
		}
		>:last-child {
			margin-left: 20px;
			flex: 1;
		}
		span {
			line-height: 1.2;
		}
	}

	.el-dialog__header {
		.username {
			display: inline-block;
			margin-left: $icon-spacing;
		}
	}

	.el-form {
		.el-form-item {
			margin-bottom: 30px;
			>.el-form-item__label {
				font-size: 1em;
				line-height: 1.2em;
			}
			>.el-form-item__content {
				>*:not(:first-child) {
					margin-top: 10px;
				}
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
			.section-heading {
				font-size: larger;
			}
			.el-select {
				margin: 0;
				width: 100%;
			}
		}
	}

	.options {
		.email {
			text-decoration: underline;
		}
		.el-button {
			display: block;
			margin: 0;
		}
		p + .el-button {
			margin-top: 20px;
		}
		.el-button + .el-button {
			margin-top: 10px;
		}
	}

}
</style>
