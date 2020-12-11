import $ from 'jquery';
import Vue from 'vue';
import Vuex from 'vuex';

import {defaultUserSettings} from './utils.js';

Vue.use(Vuex);

const $window = $(window);
const MOBILE_MAX_WIDTH = 767;
const MEDIUM_MAX_WIDTH = 991;

export const store = new Vuex.Store({
	state: {
		currentTime: Date.now(), // updated every minute
		windowWidth: $window.width(),
		dragging: false,
		movingBlockId: null, // id of block being moved
		savedScrollPosition: null, // set when returning to routes in history
		currentSpecId: null,
		cachedSpec: null,
		cachedSubspecsById: {},
		currentSpecScrollTop: null, // saved for navigation improvements
		userSettings: window.user.settings,
	},
	getters: {
		userID(state) {
			return window.user.id;
		},
		username(state) {
			return window.user.username;
		},
		userSettings(state) {
			return $.extend(true, defaultUserSettings(), state.userSettings);
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
		getCachedFullSpec(state) {
			return (id) => {
				id = parseInt(id, 10);
				if (state.cachedSpec && state.cachedSpec.id === id) {
					return state.cachedSpec;
				}
				return null;
			};
		},
		getCachedFullSubspec(state) {
			return (id) => {
				id = parseInt(id, 10);
				if (state.cachedSubspecsById[id]) {
					return state.cachedSubspecsById[id];
				}
				return null;
			};
		},
	},
	mutations: {
		setWindowWidth(state, width) {
			state.windowWidth = width;
		},
		setUserSettings(state, settings) {
			state.userSettings = settings;
		},
		startDragging(state) {
			state.dragging = true;
		},
		endDragging(state) {
			state.dragging = false;
		},
		startMovingBlock(state, blockId) {
			state.movingBlockId = blockId;
		},
		endMovingBlock(state) {
			state.movingBlockId = null;
		},
		setSavedScrollPosition(state, position) {
			console.debug('setSavedScrollPosition', position);
			state.savedScrollPosition = position;
		},
		clearSavedScrollPosition(state) {
			console.debug('clearSavedScrollPosition');
			state.savedScrollPosition = null;
		},
		saveCurrentSpecScrollTop(state, specId) {
			console.debug('saveCurrentSpecScrollTop');
			state.currentSpecId = specId;
			state.currentSpecScrollTop = $window.scrollTop();
		},
		cacheLatestFullSpec(state, spec) {
			state.cachedSpec = spec;
		},
		cacheLatestFullSubspec(state, subspec) {
			state.cachedSubspecsById[subspec.id] = subspec;
		},
		forgetSubspec(state, subspecId) {
			delete(state.cachedSubspecsById[subspecId]);
		},
		clearCurrentSpec(state) {
			state.currentSpecId = null;
			state.currentSpecScrollTop = null;
			state.cachedSpec = null;
			state.cachedSubspecsById = {};
		},
		updateCurrentTime(state) {
			state.currentTime = Date.now();
		},
	},
});

export default store;

$window.on('resize', () => {
	store.commit('setWindowWidth', $window.width());
});

const TIMEOUT = 60 * 1000;
function updateCurrentTime() {
	store.commit('updateCurrentTime');
	setTimeout(updateCurrentTime, TIMEOUT);
}
setTimeout(updateCurrentTime, TIMEOUT);
