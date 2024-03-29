<template>
<div class="admin-page">

	<header>
		<h2>Admin panel</h2>
	</header>

	<div class="content-page">

		<p v-if="loading">
			<loading-message message="Loading..."/>
		</p>
		<el-alert v-else-if="error === 403 || !$store.getters.userIsAdmin"
			type="info" show-icon :closable="false">
			On this page an administrator can manage signup requests and user accounts.
		</el-alert>
		<p v-else-if="error">Error {{error}}</p>

		<template v-else>

			<section v-if="!loading && !error" class="actions">
				<el-button @click="openEmailAll()">Email all users</el-button>
			</section>

			<section v-if="signupRequests" class="signup-requests">
				<h3>Signup requests</h3>
				<div class="options">
					<el-checkbox v-model="showAllSignupRequests">Show all</el-checkbox>
				</div>
				<p v-if="loadingSignupRequests">
					<loading-message message="Loading..."/>
				</p>
				<el-table
					v-else-if="signupRequests.length"
					:data="signupRequests"
					max-height="80vh">
					<el-table-column fixed prop="id" label="ID" width="40"/>
					<el-table-column fixed prop="username" label="Username" width="150"/>
					<el-table-column prop="email" label="Email address" width="300"/>
					<el-table-column label="Created" width="200">
						<template #default="scope">
							<moment :datetime="scope.row.created" :offset="true"/>
						</template>
					</el-table-column>
					<template v-if="showAllSignupRequests">
						<el-table-column label="Status" width="120">
							<template #default="scope">
								<template v-if="scope.row.reviewed">
									<el-tag v-if="scope.row.approved" type="success">Approved</el-tag>
									<el-tag v-else type="warning">Denied</el-tag>
								</template>
								<el-tag v-else>Pending</el-tag>
							</template>
						</el-table-column>
						<el-table-column prop="userId" label="User ID" width="120"/>
					</template>
					<el-table-column label="Actions" width="200">
						<template #default="scope">
							<el-button v-if="!scope.row.reviewed"
								@click="openReviewSignupRequest(scope.row)">
								Review
							</el-button>
						</template>
					</el-table-column>
				</el-table>
				<p v-else-if="showAllSignupRequests">No data</p>
				<p v-else>No pending requests</p>
			</section>

			<section v-if="users" class="users">
				<h3>Users</h3>
				<p v-if="loadingUsers">
					<loading-message message="Loading..."/>
				</p>
				<el-table
					v-else
					:data="users"
					max-height="80vh">
					<el-table-column fixed label="Username" width="190">
						<template #default="scope">
							<username :username="scope.row.username" :highlight="scope.row.highlight"/>
						</template>
					</el-table-column>
					<el-table-column prop="email" label="Email address" width="300"/>
					<el-table-column label="Created" width="200">
						<template #default="scope">
							<moment :datetime="scope.row.created" :offset="true"/>
						</template>
					</el-table-column>
					<el-table-column prop="specs" label="Spec count" width="100"/>
					<el-table-column label="Actions" width="200">
						<template #default="scope">
							<el-button @click="openEmailUser(scope.row.id)">Email</el-button>
						</template>
					</el-table-column>
				</el-table>
			</section>

		</template>

	</div>

	<el-dialog
		v-if="$store.getters.userIsAdmin"
		v-model="showingReviewSignupRequest"
		title="Review signup request"
		:close-on-click-modal="!sendingSignupRequestReview"
		:width="$store.getters.dialogTinyWidth"
		@closed="reviewSignupRequestModalClosed()">

		<table v-if="reviewingSignupRequest">
			<tr>
				<td>Request ID&emsp;</td>
				<td>{{reviewingSignupRequest.id}}</td>
			</tr>
			<tr>
				<td>Created&emsp;</td>
				<td><moment :datetime="reviewingSignupRequest.created"/></td>
			</tr>
			<tr>
				<td>Username&emsp;</td>
				<td><strong>{{reviewingSignupRequest.username}}</strong></td>
			</tr>
			<tr>
				<td>Email address&emsp;</td>
				<td><strong>{{reviewingSignupRequest.email}}</strong></td>
			</tr>
		</table>

		<br/><br/>

		<label>
			<div>Message</div>
			<el-input
				type="textarea"
				v-model="signupRequestResponse"
				:autosize="{minRows: 2}"
				:readonly="sendingSignupRequestReview"
				/>
		</label>

		<p v-if="sendingSignupRequestReview">
			<loading-message message="Sending..."/>
		</p>

		<template #footer>
			<span class="dialog-footer">
				<el-button
					@click="cancelReviewSignupRequest()"
					:disabled="sendingSignupRequestReview">
					Cancel
				</el-button>
				<el-button
					type="warning"
					@click="submitSignupRequestReview(false)"
					:disabled="sendingSignupRequestReview">
					Decline
				</el-button>
				<el-button
					type="primary"
					@click="submitSignupRequestReview(true)"
					:disabled="sendingSignupRequestReview">
					Approve
				</el-button>
			</span>
		</template>

	</el-dialog>

	<el-dialog
		v-if="$store.getters.userIsAdmin"
		v-model="showingSendEmailModal"
		title="Send email"
		:close-on-click-modal="false"
		:close-on-press-escape="!sendingEmail"
		:width="$store.getters.dialogSmallWidth"
		@closed="sendEmailModalClosed()">

		<label>
			<div>User</div>
			<el-select v-model="sendEmailUserId">
				<el-option value="" label="All users"/>
				<template v-if="users">
					<el-option v-for="u in users"
						:key="u.id"
						:value="u.id"
						:label="u.username + ' (' + u.email + ')'"/>
				</template>
			</el-select>
		</label>

		<label>
			<div>Subject</div>
			<el-input
				type="text"
				v-model="sendEmailSubject"
				:readonly="sendingEmail"
				/>
		</label>

		<label>
			<div>Body</div>
			<el-input
				type="textarea"
				v-model="sendEmailBody"
				:autosize="{minRows: 2}"
				:readonly="sendingEmail"
				/>
		</label>

		<p v-if="sendingEmail">
			<loading-message message="Sending..."/>
		</p>

		<template #footer>
			<span class="dialog-footer">
				<el-button
					@click="cancelSendEmail()"
					:disabled="sendingEmail">
					Cancel
				</el-button>
				<el-button
					type="primary"
					@click="sendEmail()"
					:disabled="disableSendEmail">
					Send
				</el-button>
			</span>
		</template>

	</el-dialog>

</div>
</template>

<script>
import Moment from '../widgets/moment.vue';
import Username from '../widgets/username.vue';
import LoadingMessage from '../widgets/loading.vue';
import {
	idsEq,
	alertError,
	notifySuccess,
} from '../utils.js';

export default {
	components: {
		Moment,
		Username,
		LoadingMessage,
	},
	data() {
		return {
			error: null,

			loadingSignupRequests: true,
			showAllSignupRequests: false,
			signupRequests: null,

			reviewingSignupRequest: null,
			showingReviewSignupRequest: false,
			signupRequestResponse: '',
			sendingSignupRequestReview: false,

			loadingUsers: true,
			users: null,

			showingSendEmailModal: false,
			sendEmailUserId: '',
			sendEmailSubject: '',
			sendEmailBody: '',
			sendingEmail: false,
		};
	},
	computed: {
		loading() {
			return this.loadingSignupRequests || this.loadingUsers;
		},
		disableSendEmail() {
			return !this.sendEmailSubject.trim() || !this.sendEmailBody.trim() ||
				this.sendingEmail;
		},
	},
	watch: {
		showAllSignupRequests() {
			this.loadSignupRequests();
		},
	},
	beforeRouteEnter(to, from, next) {
		next(vm => {
			vm.loadAdmin();
		});
	},
	beforeRouteUpdate(to, from, next) {
		this.loadAdmin();
		next();
	},
	methods: {
		loadAdmin() {
			this.loadSignupRequests();
			this.loadUsers();
			this.error = null;
		},
		loadSignupRequests() {
			this.loadingSignupRequests = true;
			$.get('/ajax/admin/signup-requests', {
				all: this.showAllSignupRequests,
			}).then(requests => {
				this.loadingSignupRequests = false;
				this.signupRequests = requests;
			}).fail(jqXHR => {
				this.loadingSignupRequests = false;
				this.signupRequests = [];
				this.error = jqXHR.status;
				console.error(jqXHR);
			});
		},
		loadUsers() {
			this.loadingUsers = true;
			$.get('/ajax/admin/users').then(users => {
				this.loadingUsers = false;
				this.users = users;
			}).fail(jqXHR => {
				this.loadingUsers = false;
				this.users = [];
				this.error = jqXHR.status;
				console.error(jqXHR);
			});
		},

		openReviewSignupRequest(request) {
			this.reviewingSignupRequest = request;
			this.showingReviewSignupRequest = true;
		},
		cancelReviewSignupRequest() {
			this.showingReviewSignupRequest = false;
		},
		submitSignupRequestReview(approve) {
			if (!this.reviewingSignupRequest) {
				return;
			}
			this.sendingSignupRequestReview = true;
			$.post('/ajax/admin/review-signup', {
				requestId: this.reviewingSignupRequest.id,
				approved: approve,
				message: this.signupRequestResponse,
			}).then(() => {
				this.sendingSignupRequestReview = false;
				for (var i = 0; i < this.signupRequests.length; i++) {
					if (idsEq(this.signupRequests[i].id, this.reviewingSignupRequest.id)) {
						if (this.showAllSignupRequests) {
							this.reviewingSignupRequest.reviewed = true;
							this.reviewingSignupRequest.approved = approve;
							this.signupRequests.splice(i, 1, this.reviewingSignupRequest); // Replace in array
						} else {
							this.signupRequests.splice(i, 1); // Remove from array
						}
						break;
					}
				}
				this.showingReviewSignupRequest = false;
			}).fail(jqXHR => {
				this.sendingSignupRequestReview = false;
				alertError(jqXHR);
			});
		},
		reviewSignupRequestModalClosed() {
			this.reviewingSignupRequest = null;
			this.signupRequestResponse = '';
		},

		openEmailAll() {
			this.sendEmailUserId = '';
			this.showingSendEmailModal = true;
		},
		openEmailUser(userId) {
			this.sendEmailUserId = userId;
			this.showingSendEmailModal = true;
		},
		cancelSendEmail() {
			this.showingSendEmailModal = false;
		},
		sendEmail() {
			this.sendingEmail = true;
			$.post('/ajax/admin/send-email', {
				userId: this.sendEmailUserId,
				subject: this.sendEmailSubject.trim(),
				body: this.sendEmailBody.trim(),
			}).then(() => {
				this.sendingEmail = false;
				this.showingSendEmailModal = false;
				notifySuccess('Email sent');
			}).fail(jqXHR => {
				this.sendingEmail = false;
				alertError(jqXHR);
			});
		},
		sendEmailModalClosed() {
			this.sendEmailUserId = '';
			this.sendEmailSubject = '';
			this.sendEmailBody = '';
		},
	},
};
</script>

<style lang="scss">
@import '../_styles/_breakpoints.scss';
@import '../_styles/_colours.scss';
@import '../_styles/_app.scss';

.admin-page {
	>header {
		background-color: $admin-bg;
		color: white;
		// overflow: hidden; // keep {float: right} content bounded on mobile

		padding: $page-header-vertical-padding $page-header-horiz-padding;
		@include mobile {
			padding: $page-header-vertical-padding-sm $page-header-horiz-padding-sm;
		}

		>h2 {
			margin: 0;
		}
	}

	>.content-page {
		>section {
			margin: 60px 0;
			&:first-child {
				margin-top: 0;
			}
			>.options {
				margin-bottom: 40px;
			}
		}
	}

	.el-textarea {
		width: 100%;
	}

	.el-dialog__body {
		>label {
			display: block;
			>div:first-child:not([class]) {
				margin-bottom: 5px;
			}
			>.el-select {
				width: 100%;
			}
		}
		>*+* {
			margin-top: 20px;
		}
	}
}
</style>
