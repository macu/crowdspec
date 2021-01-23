import $ from 'jquery';

const $window = $(window);

export default new Vue({
	data() {
		return {
			width: $window.width(),
			height: $window.height(),
			scrollTop: $window.scrollTop(),
			scrollLeft: $window.scrollLeft(),
		};
	},
	computed: {
		dims() {
			return {
				width: this.width,
				height: this.height,
				scrollTop: this.scrollTop,
				scrollLeft: this.scrollLeft,
			};
		},
	},
	created() {
		this.onResize = $window.on('resize', this.onResize);
		this.onScroll = $window.on('scroll', this.onScroll);
	},
	beforeDestroy() {
		$window.off('resize', this.onResize);
		$window.off('scroll', this.onScroll);
	},
	methods: {
		onResize() {
			this.width = $window.width();
			this.height = $window.height();
		},
		onScroll() {
			this.scrollTop = $window.scrollTop();
			this.scrollLeft = $window.scrollLeft();
		},
		testElementWithinViewport(el, dims = this) {
			let $el = $(el),
				offset = $el.offset(),
				width = $el.outerWidth(),
				height = $el.outerHeight();
			// Test whether any part of the element is within the viewport
			return (
				// top edge above botton of viewport
				(offset.top < (dims.scrollTop + dims.height)) &&
				// bottom edge below top of viewport
				((offset.top + height) > dims.scrollTop) &&
				// left edge to left of right edge of viewport
				(offset.left < (dims.scrollLeft + dims.width)) &&
				// right edge to right of left edge of viewport
				((offset.left + width) > dims.scrollLeft)
			);
		},
	},
});
