import $ from 'jquery';
import Vue from 'vue';
import Vuex from 'vuex';

Vue.use(Vuex);

const $window = $(window);
const MOBILE_MAX_WIDTH = 767;
const MEDIUM_MAX_WIDTH = 991;

export const store = new Vuex.Store({
	state: {
		windowWidth: $window.width(),
	},
	getters: {
		userID(state) {
			return window.user.id;
		},
		username(state) {
			return window.user.username;
		},
		dialogTinyWidth(state) {
			if (state.windowWidth <= MOBILE_MAX_WIDTH) {
				return '80%';
			} else if (state.windowWidth <= MEDIUM_MAX_WIDTH) {
				return '50%';
			} else {
				return '30%';
			}
		},
		dialogSmallWidth(state) {
			if (state.windowWidth <= MOBILE_MAX_WIDTH) {
				return '90%';
			} else if (state.windowWidth <= MEDIUM_MAX_WIDTH) {
				return '75%';
			} else {
				return '50%';
			}
		},
		dialogLargeWidth(state) {
			if (state.windowWidth <= MEDIUM_MAX_WIDTH) {
				return '95%';
			} else {
				return '90%';
			}
		},
	},
	mutations: {
		setWindowWidth(state, width) {
			state.windowWidth = width;
		},
	},
});

export default store;

$window.on('resize', () => {
	store.commit('setWindowWidth', $window.width());
});
