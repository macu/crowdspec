<template>
<div class="app">
	<div class="header">
		<h1 @click="gotoIndex()">CrowdSpec</h1>
		<span>{{username}}</span>
		<el-button @click="logout()" size="mini">Log out</el-button>
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
@import './styles/_colours.scss';

.app {
	height: 100%;
	>.header {
		display: flex;
		padding: 20px 1cm;
		align-items: center;
		background-color: $app;
		color: white;
		>h1 {
			flex: 1;
			margin: 0;
			cursor: pointer;
		}
		>*:not(:last-child) {
			margin-right: 20px;
		}
	}
	>.content-area {
		padding: 1cm;
	}
}
</style>
