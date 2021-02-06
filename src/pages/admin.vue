<template>
<div class="admin-page">

	<header>
		<h2>Admin panel</h2>
	</header>

	<div class="content-page">

		<p v-if="loading">Loading...</p>
		<p v-else-if="error === 403">Unauthorized</p>
		<p v-else-if="error">Error {{error}}</p>

		<section v-if="signupRequests" class="signup-requests">
			<h3>Signup requests</h3>
			<div class="options">
				<el-checkbox v-model="showAllSignupRequests">Show all</el-checkbox>
			</div>
			<p v-if="loadingSignupRequests">Loading...</p>
			<el-table
				v-else-if="signupRequests.length"
				:data="signupRequests"
				:max-height=".80 * $store.state.windowHeight"
				style="width:100%;">
				<el-table-column fixed prop="id" label="ID" width="40"/>
				<el-table-column fixed prop="username" label="Username" width="150"/>
				<el-table-column prop="email" label="Email address" width="300"/>
				<el-table-column label="Created" width="200">
					<template slot-scope="scope">
						<moment :datetime="scope.row.created" :offset="true"/>
					</template>
				</el-table-column>
				<el-table-column label="Actions" width="200">
					<template slot-scope="scope">
						<el-button v-if="!scope.row.reviewed"
							@click="openReviewSignupRequest(scope.row)"
							size="mini">
							Review
						</el-button>
					</template>
				</el-table-column>
				<template v-if="showAllSignupRequests">
					<el-table-column label="Reviewed" width="120">
						<template slot-scope="scope">
							<el-tag v-if="scope.row.reviewed" type="success">Yes</el-tag>
							<el-tag v-else type="warning">No</el-tag>
						</template>
					</el-table-column>
					<el-table-column label="Approved" width="120">
						<template slot-scope="scope">
							<template v-if="scope.row.reviewed">
								<el-tag v-if="scope.row.approved" type="success">Yes</el-tag>
								<el-tag v-else type="warning">No</el-tag>
							</template>
						</template>
					</el-table-column>
					<el-table-column prop="userId" label="User ID" width="120"/>
				</template>
			</el-table>
			<p v-else-if="showAllSignupRequests">No data</p>
			<p v-else>No pending requests</p>
		</section>

	</div>

	<el-dialog
		v-if="reviewingSignupRequest"
		title="Review signup request"
		:close-on-click-modal="!sendingSignupRequestReview"
		:width="$store.getters.dialogTinyWidth"
		visible>

		<table>
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
			Message
			<el-input
				type="textarea"
				v-model="signupRequestResponse"
				:autosize="{minRows: 2}"
				:readonly="sendingSignupRequestReview"
				/>
		</label>

		<p v-if="sendingSignupRequestReview">
			<i class="el-icon-loading"/> Sending...
		</p>

		<span slot="footer" class="dialog-footer">
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

	</el-dialog>

</div>
</template>

<script>
import Moment from '../widgets/moment.vue';
import {idsEq, alertError} from '../utils.js';

export default {
	components: {
		Moment,
	},
	data() {
		return {
			loadingSignupRequests: true,
			error: null,
			showAllSignupRequests: false,
			signupRequests: null,
			reviewingSignupRequest: null,
			signupRequestResponse: '',
			sendingSignupRequestReview: false,
		};
	},
	computed: {
		loading() {
			return this.loadingSignupRequests;
		},
		unauthorized() {
			return !(this.loading || this.error || this.signupRequests);
		},
	},
	watch: {
		showAllSignupRequests() {
			this.loadAdmin();
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
			this.loadingSignupRequests = true;
			this.error = null;
			$.get('/ajax/admin/signup-requests', {
				all: this.showAllSignupRequests,
			}).then(requests => {
				this.loadingSignupRequests = false;
				this.signupRequests = requests;
			}).fail(jqXHR => {
				this.loadingSignupRequests = false;
				this.error = jqXHR.status;
				alertError(jqXHR);
			});
		},
		openReviewSignupRequest(request) {
			this.signupRequestResponse = '';
			this.reviewingSignupRequest = request;
		},
		cancelReviewSignupRequest() {
			this.reviewingSignupRequest = null;
			this.signupRequestResponse = '';
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
				this.reviewingSignupRequest = null;
				this.signupRequestResponse = '';
			}).fail(jqXHR => {
				this.sendingSignupRequestReview = false;
				alertError(jqXHR);
			});
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
			margin-bottom: 60px;
			>.options {
				margin-bottom: 40px;
			}
		}
	}

	.el-textarea {
		width: 100%;
	}
}
</style>
