<template>
<div class="app">
	<div class="header">
		<h1 @click="gotoIndex()">CrowdSpec</h1>
		<div>
			<span>{{username}}</span>
			<el-button @click="logout()" size="mini">Log out</el-button>
		</div>
	</div>
	<div class="content-area">
		<router-view></router-view>
	</div>
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

.app {
	height: 100%;

	>.header {
		display: flex;
		padding: 20px 8mm 10px;
		align-items: center;
		justify-content: flex-end;
		background-color: $app;
		color: white;

		flex-wrap: wrap;

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

	>.content-area {
		padding: 1cm;
	}
}
</style>
