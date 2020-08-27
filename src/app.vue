<template>
<div class="app">
	<header>
		<h1 @click="gotoIndex()">CrowdSpec</h1>
		<div>
			<span>{{username}}</span>
			<el-button @click="logout()" size="mini">Log out</el-button>
		</div>
	</header>
	<router-view class="page-area"/>
</div>
</template>

<script>
import $ from 'jquery';
import store from './store.js';
import router from './router.js';

export default {
	store,
	router,
	computed: {
		username() {
			return this.$store.getters.username;
		},
	},
	methods: {
		gotoIndex() {
			if (this.$route.name !== 'index') {
				this.$router.push({name: 'index'});
			}
		},
		logout() {
			window.location.href = '/logout';
		},
	},
};
</script>

<style lang="scss">
@import './styles/_breakpoints.scss';
@import './styles/_colours.scss';
@import './styles/_app.scss';

.app {
	height: 100%;

	>header {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		flex-wrap: wrap;
		background-color: $app-bg;
		color: white;

		padding: 20px $app-header-horiz-padding 10px;
		@include mobile {
			padding: 15px $app-header-horiz-padding-sm 5px;
		}

		>h1 {
			flex: 1;
			margin: 0 0 10px;
			padding-right: 20px;
			cursor: pointer;
		}

		>div {
			margin: 0 0 10px;
			>*:not(:first-child) {
				margin-left: 20px;
			}
		}
	}

	>.page-area {
		padding: $content-area-padding;
		>header {
			margin-top: #{-$content-area-padding};
			margin-left: #{-$content-area-padding};
			margin-right: #{-$content-area-padding};
		}
		@include mobile {
			padding: $content-area-padding-sm;
			>header {
				margin-top: #{-$content-area-padding-sm};
				margin-left: #{-$content-area-padding-sm};
				margin-right: #{-$content-area-padding-sm};
			}
		}
	}
}
</style>
