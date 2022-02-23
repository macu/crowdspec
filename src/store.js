import $ from 'jquery';
import {createStore} from 'vuex';

import {idsEq, defaultUserSettings} from './utils.js';
import {OWNER_TYPE_USER} from './spec/const.js';

const $window = $(window);
const MOBILE_MAX_WIDTH = 767;
const MEDIUM_MAX_WIDTH = 991;

export const store = Vuex.createStore({
	state() {
		return {
			currentTime: Date.now(), // updated every minute
			windowWidth: $window.width(),
			windowHeight: $window.height(),
			dragging: false, // drag operation in progress; lock down interface
			movingBlockIds: [], // ids of blocks being moved
			movingBlocksSourceSubspecId: null, // id of source subspec if any for current move operation
			savedScrollPosition: null, // set momentarily when returning to routes in history
			currentSpec: null,
			currentSpecScrollTop: null, // saved when leaving spec view for navigation improvements
			userSettings: window.user.settings,
		};
	},
	getters: {
		currentUserId(state) {
			return window.user.id;
		},
		username(state) {
			return window.user.username;
		},
		emailAddress(state) {
			return window.user.email;
		},
		userSettings(state) {
			return $.extend(true, defaultUserSettings(), state.userSettings);
		},
		userIsAdmin(state) {
			return window.user.admin;
		},
		mobileViewport(state) {
			return state.windowWidth <= MOBILE_MAX_WIDTH;
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
		currentlyMovingBlocks(state) {
			return state.movingBlockIds.length > 0;
		},
		currentlyMovingBlock(state) {
			return (blockId) => {
				return state.movingBlockIds.indexOf(blockId) >= 0;
			};
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
		updateWindowDimensions(state) {
			state.windowWidth = $window.width();
			state.windowHeight = $window.height();
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
		setMovingBlocks(state, payload) {
			state.movingBlocksSourceSubspecId = payload.subspecId;
			state.movingBlockIds = payload.blockIds;
		},
		endMovingBlocks(state) {
			state.movingBlocksSourceSubspecId = null;
			state.movingBlockIds = [];
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
	store.commit('updateWindowDimensions');
});

const TIMEOUT = 60 * 1000;
function updateCurrentTime() {
	store.commit('updateCurrentTime');
	setTimeout(updateCurrentTime, TIMEOUT);
}
setTimeout(updateCurrentTime, TIMEOUT);
