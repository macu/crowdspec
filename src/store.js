import $ from 'jquery';
import Vue from 'vue';
import Vuex from 'vuex';

import {idsEq, defaultUserSettings} from './utils.js';
import {OWNER_TYPE_USER} from './spec/const.js';

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
		currentSpec: null,
		currentSpecScrollTop: null, // saved for navigation improvements
		userSettings: window.user.settings,
	},
	getters: {
		currentUserId(state) {
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
				return '30%'; // TODO set pixel width
			}
		},
		dialogSmallWidth(state) {
			if (state.windowWidth <= MOBILE_MAX_WIDTH) {
				return '90%';
			} else if (state.windowWidth <= MEDIUM_MAX_WIDTH) {
				return '75%';
			} else {
				return '50%'; // TODO set pixel width
			}
		},
		dialogLargeWidth(state) {
			if (state.windowWidth <= MEDIUM_MAX_WIDTH) {
				return '95%';
			} else {
				return '90%'; // TODO set pixel width
			}
		},
		currentSpecId(state) {
			return state.currentSpec ? state.currentSpec.id : null;
		},
		currentSpecOwnedByUser(state, getters) {
			if (state.currentSpec) {
				return state.currentSpec.ownerType === OWNER_TYPE_USER &&
					idsEq(state.currentSpec.ownerId, getters.currentUserId);
			}
			return false;
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
		saveCurrentSpec(state, spec) {
			console.debug('saveCurrentSpec');
			if (state.currentSpec && !idsEq(state.currentSpec.id, spec.id)) {
				console.debug('saveCurrentSpec clear currentSpecScrollTop');
				state.currentSpecScrollTop = null;
			}
			state.currentSpec = spec;
		},
		saveCurrentSpecScrollTop(state) {
			console.debug('saveCurrentSpecScrollTop');
			state.currentSpecScrollTop = $window.scrollTop();
		},
		clearCurrentSpec(state) {
			state.currentSpec = null;
			state.currentSpecScrollTop = null;
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
