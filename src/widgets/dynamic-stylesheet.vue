<template>
	<span class="stylesheet-placeholder"></span>
</template>

<script>
export default {
	props: {
		/* Accepts values like
			{
				'div': {
					'width': '20px',
					'height': '20px',
					'background-color': 'blue',
					'color': 'white',

					'>span': {
						'color': 'limegreen',
					},
				},
			} */
		rules: Object,
		innerMarkupOutput: Boolean,
	},
	watch: {
		rules: {
			deep: true,
			handler() {
				this.recreateSheet();
			},
		},
	},
	mounted() {
		this.recreateSheet();
	},
	beforeDestroy() {
		this.removeSheet();
	},
	methods: {
		recreateSheet() {
			// use stylesheet API if available,
			// otherwise create a new sheet using innerHTML

			// TODO handle media queries

			if (this.$stylesheet) {
				document.head.removeChild(this.$stylesheet);
				this.$stylesheet = null;
			}

			if (!this.rules) {
				return;
			}

			let style = document.createElement('style');
			style.type = 'text/css';

			document.head.appendChild(style);
			this.$stylesheet = style;

			const addRules = (rules, selectors = [], index = 0) => {
				if (!rules || typeof rules !== 'object') {
					return 0;
				}

				let added = 0;

				let stringRules = [];
				const addStringRules = () => {
					for (let i = 0; i < stringRules.length; i++) {
						let rule = stringRules[i].trim();
						if (!rule) {
							continue;
						}
						if (rule.charAt(rule.length - 1) !== ';') {
							stringRules[i] = rule + ';';
						}
					}
					stringRules = stringRules.join('');
					if (stringRules) {
						let selector = selectors.join('').trim();
						// Thanks https://stackoverflow.com/a/22697964/1597274
						if (!(style.sheet||{}).insertRule) {
							(style.styleSheet || style.sheet).addRule(selector, stringRules, index + added);
						} else {
							style.sheet.insertRule(selector + '{' + stringRules + '}', index + added);
						}
						added++;
					}
					stringRules = [];
				};

				if (Array.isArray(rules)) {
					for (let i = 0; i < rules.length; i++) {
						let rule = rules[i];
						if (typeof rule === 'string') { // take as literal name:value rule
							stringRules.push(rule);
						} else if (typeof rule === 'object') { // object or array
							addStringRules();
							added += addRules(rule, selectors, index + added);
						}
					}
				} else {
					for (let key in rules) {
						if (rules.hasOwnProperty(key)) {
							let rule = rules[key];
							if (typeof rule === 'string') {
								stringRules.push(key + ':' + rule + ';');
							} else if (typeof rule === 'object') { // object or string
								addStringRules();
								if (['>', ':', '+', '~', ' '].indexOf(key.charAt(0)) < 0) {
									key = ' ' + key; // descendent selector
								}
								added += addRules(rule, selectors.concat([key]), index + added);
							}
						}
					}
				}

				addStringRules();

				return added;
			};

			const buildCSS = (rules, selectors = []) => {
				if (!rules || typeof rules !== 'object') {
					return 0;
				}

				let added = '';

				let stringRules = [];
				const addStringRules = () => {
					let compiledStringRules = [];
					for (let i = 0; i < stringRules.length; i++) {
						let rule = stringRules[i].trim();
						if (!rule) {
							// Remove blank rules from output
							continue;
						}
						if (rule.charAt(rule.length - 1) === ';') {
							compiledStringRules.push(rule);
						} else {
							compiledStringRules.push(rule + ';');
						}
					}
					if (compiledStringRules.length) {
						compiledStringRules = compiledStringRules.join('\n\t');
						let selector = selectors.join('').trim();
						added += '\n' + selector + ' {\n\t' + compiledStringRules + '\n}';
					}
					stringRules = [];
				};

				if (Array.isArray(rules)) {
					for (let i = 0; i < rules.length; i++) {
						let rule = rules[i];
						if (typeof rule === 'string') { // take as literal name:value rule
							stringRules.push(rule);
						} else if (typeof rule === 'object') { // object or array
							addStringRules();
							added += buildCSS(rule, selectors);
						}
					}
				} else {
					for (let key in rules) {
						if (rules.hasOwnProperty(key)) {
							let rule = rules[key];
							if (typeof rule === 'string') {
								stringRules.push(key + ': ' + rule + ';');
							} else if (typeof rule === 'object') { // object or string
								addStringRules();
								if (['>', ':', '+', '~', ' '].indexOf(key.charAt(0)) < 0) {
									key = ' ' + key; // descendent selector
								}
								added += buildCSS(rule, selectors.concat([key]));
							}
						}
					}
				}

				addStringRules();

				return added + (selectors.length ? '' : '\n');
			};

			if ((style.styleSheet || style.sheet) && !this.innerMarkupOutput) {
				addRules(this.rules);
			} else {
				console.debug('dynamic-stylesheet: rendering innerHTML stylesheet');
				style.innerHTML = buildCSS(this.rules);
			}
		},
		removeSheet() {
			if (this.$stylesheet) {
				document.head.removeChild(this.$stylesheet);
				this.$stylesheet = null;
			}
		},
	},
};
</script>

<style lang="scss">
.stylesheet-placeholder {
	display: none;
}
</style>
