<template>
<div class="app">

	<header>
		<h1 @click="gotoIndex()">CrowdSpec</h1>
		<div v-if="loadingUser">
			<span>Loading...</span>
		</div>
		<div v-else-if="loggedIn">
			<el-button @click="openEditProfile()" type="text" class="username-button">
				<username :username="username" :highlight="highlight"/>
			</el-button>
			<el-button @click="logout()">Log out</el-button>
		</div>
		<div v-else>
			<el-button @click="gotoLogin()">Log in</el-button>
			<el-button @click="gotoSignup()">Sign up</el-button>
		</div>
	</header>

	<router-view class="page-area"/>

	<edit-profile-modal v-if="loggedIn" ref="editProfileModal"/>

</div>
</template>

<script>
import Username from './widgets/username.vue';
import EditProfileModal from './widgets/edit-profile-modal.vue';
import {promptConfirm} from './utils.js';

export default {
	components: {
		Username,
		EditProfileModal,
	},
	computed: {
		loadingUser() {
			return this.$store.getters.loadingUser;
		},
		loggedIn() {
			return this.$store.getters.loggedIn;
		},
		username() {
			return this.$store.getters.username;
		},
		highlight() {
			return this.$store.getters.userSettings.userProfile.highlightUsername;
		},
	},
	mounted() {
		this.$store.dispatch('loadAuth');
	},
	methods: {
		gotoLogin() {
			window.location.href = '/login';
		},
		gotoSignup() {
			window.location.href = '/signup';
		},
		gotoIndex() {
			if (this.$route.name !== 'index') {
				this.$router.push({name: 'index'});
			}
		},
		logout() {
			promptConfirm(
				null,
				'Please confirm that you wish to log out.',
				'Log out',
			).then(() => {
				this.$store.dispatch('logOut');
			}).catch(() => {
				// Cancelled
			});
		},
		openEditProfile() {
			this.$refs.editProfileModal.show();
		},
	},
};
</script>

<style lang="scss">
@use "sass:math";
@import './_styles/_breakpoints.scss';
@import './_styles/_colours.scss';
@import './_styles/_app.scss';

.app {
	height: 100%;

	>header {
		display: flex;
		flex-direction: row; // horizontal
		align-items: center; // vertical align
		justify-content: flex-end; // align right
		flex-wrap: wrap;
		background-color: $app-bg;
		color: white;

		padding: 20px $app-header-horiz-padding 10px;
		@include mobile {
			padding: 15px $app-header-horiz-padding-sm 5px;
		}

		>h1 {
			flex: 1; // claim all extra horizontal space on line
			margin: 0 0 10px;
			padding-right: 20px;
			cursor: pointer;
		}

		>div {
			margin: 0 0 10px;

			>.username-button {
				color: white;
				padding-left: 5px;
				padding-right: 5px;
				border-bottom: 1px solid white;
			}

			>*:not(:first-child) {
				margin-left: 20px;
			}
		}
	}

	.content-page { // just content - no header - standard padding
		padding: $content-area-padding;
		padding-top: math.div($content-area-padding, 2);

		@include mobile {
			padding: $content-area-padding-sm;
		}
	}
}
</style>
