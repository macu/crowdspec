import {createRouter, createWebHashHistory} from 'vue-router';
import store from './store.js';
import {idsEq} from './utils.js';

import IndexPage from './pages/index.vue';
import SpecPage from './pages/spec.vue';
import SpecViewPage from './pages/spec-view.vue';
import SubspecPage from './pages/subspec.vue';
import SubspecViewPage from './pages/subspec-view.vue';
import CommunityReviewPage from './pages/community-review.vue';
import AdminPage from './pages/admin.vue';
import AjaxErrorPage from './pages/ajax-error.vue';
import NotFoundPage from './pages/not-found.vue';

export const router = createRouter({
	history: createWebHashHistory(),
	routes: [
		{name: 'index', path: '/', component: IndexPage},
		{path: '/spec/:specId', component: SpecPage, children: [
			{name: 'spec', path: '', component: SpecViewPage},
				{path: 'subspec/:subspecId', component: SubspecPage, children: [
				{name: 'subspec', path: '', component: SubspecViewPage},
			]},
		]},
		{name: 'community-review', path: '/community-review', component: CommunityReviewPage},
		{name: 'admin', path: '/admin', component: AdminPage},
		{name: 'ajax-error', path: '/ajax-error/:code', component: AjaxErrorPage},
		{path: '/:pathMatch(.*)', component: NotFoundPage},
	],
	scrollBehavior(to, from, savedPosition) {
		// scrollBehavior is called after the new route has been rendered.
		// savedPosition is the position previously saved at the now restored position in navigation history.
		// save to allow restoring scroll position following additional DOM updates
		// made during the route mounted hook
		store.commit('setSavedScrollPosition', savedPosition);
		return savedPosition ? savedPosition : {left: 0, top: 0};
	},
});

router.beforeEach((to, from, next) => {
	// beforeEach is called before navigation is confirmed.
	if (from) {
		if (
			store.getters.currentSpecId && to.params.specId &&
			idsEq(to.params.specId, store.getters.currentSpecId)
		) {
			// Staying within spec context;
			// save scroll top to restore upon forward navigation back to spec
			store.commit('saveCurrentSpecScrollTop');
		} else {
			// Leaving spec context for another context
			store.commit('clearCurrentSpec');
			store.commit('endMovingBlocks');
		}
	}
	next();
});

router.afterEach((to, from) => {
	// afterEach is called after navigation is confirmed,
	// but before the new route has been rendered.
	// afterEach is called before scrollBehavior is called for the new route.
	// clear the savedScrollPosition restored on the previous route
	console.debug('afterEach');
	store.commit('clearSavedScrollPosition');
});

export default router;
