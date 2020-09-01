<template>
<el-dialog
	title="Edit profile"
	:visible.sync="showing"
	:width="$store.getters.dialogTinyWidth"
	:close-on-click-modal="false"
	@closed="clearMode()"
	class="edit-profile-modal">

	<p>User ID: {{userID}}</p>
	<p>Username: {{username}}</p>

	<el-form v-if="changePasswordMode"
		ref="changePasswordForm"
		:model="changePasswordForm"
		:rules="changePasswordRules"
		label-position="top">
		<el-form-item label="Existing password" prop="oldPass">
			<el-input
				type="password"
				name="old_password"
				v-model="changePasswordForm.oldPass"
				autocomplete="off"/>
		</el-form-item>
		<el-form-item label="New password" prop="newPass">
			<el-input
				type="password"
				name="new_password"
				v-model="changePasswordForm.newPass"
				autocomplete="off"/>
		</el-form-item>
		<el-form-item label="Confirm new password" prop="newPass2">
			<el-input
				type="password"
				name="verify_password"
				v-model="changePasswordForm.newPass2"
				autocomplete="off"/>
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

export default {
	data() {
		return {
			showing: false,
			mode: null,
			changePasswordFormValid: false,
			changePasswordForm: {
				oldPass: '',
				newPass: '',
				newPass2: '',
			},
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
		},
		updateChangePasswordFormValid() {
			this.$nextTick(() => {
				this.$refs.changePasswordForm.validate(valid => {
					this.changePasswordFormValid = valid;
				});
			});
		},
		submitChangePassword() {
			this.$refs.changePasswordForm.validate(valid => {
				if (valid) {
					$.post('/ajax/user/change-password', {
						old: this.changePasswordForm.oldPass,
						new: this.changePasswordForm.newPass,
						new2: this.changePasswordForm.newPass2,
					}).then(() => {
						this.$message({
							message: 'Password updated successfully',
							type: 'success',
							showClose: true,
							duration: 1200,
						});
						this.clearMode();
					}).fail(error => {
						if (error && error.readyState) {
							switch (error.status) {
								case 400: // Bad request
									this.$alert('New password rejected', {type: 'error'});
									break;
								case 403: // Forbidden
									this.$alert('Old password rejected', {type: 'error'});
									break;
								default:
									alertError(error);
							}
						} else {
							alertError(error);
						}
					});
				}
			});
		},
		clearMode() {
			this.mode = null;
			this.changePasswordFormValid = false;
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
	.el-button.change-password {
		margin-top: 40px;
	}
	.el-form {
		margin-top: 40px;
		.el-form-item {
			margin-bottom: 10px;
			.el-form-item__label {
				font-size: 1em;
				line-height: 1.2em;
			}
		}
	}
}
</style>
