import $ from 'jquery';
import {createStore} from 'vuex';
import {alertError} from './utils.js';

import {idsEq, defaultUserSettings} from './utils.js';
import {OWNER_TYPE_USER} from './spec/const.js';

const $window = $(window);
const MOBILE_MAX_WIDTH = 767;
const MEDIUM_MAX_WIDTH = 991;

let app = null;
export function registerApp(appContext) {
	app = appContext;
}
export function getAppContext() {
	return app;
}

export const store = createStore({
	state() {
		return {
			currentTime: Date.now(), // updated every minute
			windowWidth: $window.width(),
			windowHeight: $window.height(),

			currentSpec: null,
			savedScrollPosition: null, // set when returning to routes in history
			currentSpecScrollTop: null, // saved when leaving spec to restore on forward navigation
			dragging: false, // drag operation in progress; lock down interface
			movingBlockIds: [], // ids of blocks being moved
			movingBlocksSourceSubspecId: null, // id of source subspec if any for current move operation

			user: null,
			userSettings: null,
		};
	},
	getters: {

		loggedIn(state) {
			return !!state.user;
		},
		loadingUser(state) {
			return state.user === false; // false means loading
		},
		currentUserId(state) {
			return state.user ? state.user.id : null;
		},
		username(state) {
			return state.user ? state.user.username : null;
		},
		emailAddress(state) {
			return state.user ? state.user.email : null;
		},
		userSettings(state) {
			if (state.userSettings) {
				// Settings have been modified since auth
				return $.extend(true, defaultUserSettings(), state.userSettings);
			} else if (state.user) {
				return $.extend(true, defaultUserSettings(), state.user.settings);
			}
			return defaultUserSettings();
		},
		userIsAdmin(state) {
			return state.user ? state.user.admin : false;
		},
		specsLayoutList(state, getters) {
			return !getters.loggedIn ||
				getters.userSettings.homepage.specsLayout === 'list';
		},
		defaultShowUnreadCommentsOnly(state, getters) {
			return getters.loggedIn && getters.userSettings.community.unreadOnly;
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
		currentlyMovingBlocks(state) {
			return state.movingBlockIds.length > 0;
		},
		currentlyMovingBlock(state) {
			return (blockId) => {
				return state.movingBlockIds.indexOf(blockId) >= 0;
			};
		},

	},
	mutations: {

		updateWindowDimensions(state) {
			state.windowWidth = $window.width();
			state.windowHeight = $window.height();
		},
		updateCurrentTime(state) {
			state.currentTime = Date.now();
		},

		setUser(state, user) {
			state.user = user;
			state.user = user;
		},
		setUserSettings(state, settings) {
			state.userSettings = settings;
		},

		saveCurrentSpec(state, spec) {
			console.debug('saveCurrentSpec');
			if (state.currentSpec && !idsEq(state.currentSpec.id, spec.id)) {
				console.debug('saveCurrentSpec clear currentSpecScrollTop');
				state.currentSpecScrollTop = null;
			}
			state.currentSpec = spec;
		},
		startDragging(state) {
			// Mouse-powered move operation now in progress
			state.dragging = true;
		},
		endDragging(state) {
			// Mouse drop
			state.dragging = false;
		},
		setMovingBlocks(state, payload) {
			// Block multi-select precise move active
			state.movingBlocksSourceSubspecId = payload.subspecId;
			state.movingBlockIds = payload.blockIds;
		},
		endMovingBlocks(state) {
			// Cancelled or fulfilled
			state.movingBlocksSourceSubspecId = null;
			state.movingBlockIds = [];
		},
		setSavedScrollPosition(state, position) {
			// Receiving saved scroll offset from router to restore navigation
			console.debug('setSavedScrollPosition', position);
			state.savedScrollPosition = position;
		},
		clearSavedScrollPosition(state) {
			// called before scrollBehavior is called for the new route
			console.debug('clearSavedScrollPosition');
			state.savedScrollPosition = null;
		},
		saveCurrentSpecScrollTop(state) {
			// called before navigating away
			console.debug('saveCurrentSpecScrollTop');
			state.currentSpecScrollTop = $window.scrollTop();
		},
		clearCurrentSpec(state) {
			state.currentSpec = null;
			state.currentSpecScrollTop = null;
		},

	},
	actions: {
		loadAuth({commit}) {
			commit('setUser', false); // false means loading
			$.get('/ajax/auth').then(response => {
				commit('setUser', response);
			}).fail(jqXHR => {
				commit('setUser', null);
				if (jqXHR && jqXHR.readyState && jqXHR.status) {
					if (jqXHR.status === 403) {
						// Forbidden - not logged in
						return;
					}
				}
				alertError(jqXHR);
			});
		},
		logOut({commit}) {
			$.get('/ajax/logout').then(response => {
				commit('setUser', null);
			}).fail(alertError);
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
