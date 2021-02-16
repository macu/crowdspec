import $ from 'jquery';
import {alertError} from '../utils.js';

export const SCRIPT_HLJS = 'highlight';
export const SCRIPT_DRAGULA = 'dragula';

let loadedScripts = {};

function loadAssets(assets) {
	let promises = [];
	for (var i = 0; i < assets.length; i++) {
		let a = assets[i];
		if (/\.js$/.test(a.src)) {
			let script = document.createElement('script');
			script.type = 'text/javascript';
			script.crossOrigin = 'anonymous';
			if (a.integrity) {
				script.integrity = a.integrity;
			}
			let p = $.Deferred();
			script.addEventListener('load', function () {
				p.resolve();
			});
			script.addEventListener('error', function () {
				p.reject();
			});
			promises.push(p);
			script.src = a.src;
			document.body.appendChild(script);
		} else if (/\.css$/.test(a.src)) {
			let link = document.createElement('link');
			link.rel = 'stylesheet';
			link.crossOrigin = 'anonymous';
			if (a.integrity) {
				link.integrity = a.integrity;
			}
			let p = $.Deferred();
			link.addEventListener('load', function () {
				p.resolve();
			});
			link.addEventListener('error', function () {
				p.reject();
			});
			promises.push(p);
			link.href = a.src;
			document.body.appendChild(link);
		}
	}
	// Return promise pending all assets
	return $.when.apply($, promises);
}

export function loadScript(scriptKey) {
	if (loadedScripts[scriptKey]) {
		return loadedScripts[scriptKey];
	}
	switch (scriptKey) {
		case SCRIPT_HLJS:
			return loadedScripts[scriptKey] = loadAssets([
				{
					src: 'https://cdnjs.cloudflare.com/ajax/libs/highlight.js/10.6.0/styles/default.min.css',
					integrity: 'sha512-kZqGbhf9JTB4bVJ0G8HCkqmaPcRgo88F0dneK30yku5Y/dep7CZfCnNml2Je/sY4lBoqoksXz4PtVXS4GHSUzQ==',
				},
				{
					src: 'https://cdnjs.cloudflare.com/ajax/libs/highlight.js/10.6.0/highlight.min.js',
					integrity: 'sha512-zol3kFQ5tnYhL7PzGt0LnllHHVWRGt2bTCIywDiScVvLIlaDOVJ6sPdJTVi0m3rA660RT+yZxkkRzMbb1L8Zkw==',
				},
			]).then(function() {
				console.debug('highlight.js loaded');
				// Resolve library reference
				return window.hljs;
			}).fail(function() {
				alertError('Failed to load highlight.js code highlighting library');
			});
		case SCRIPT_DRAGULA:
			return loadedScripts[scriptKey] = loadAssets([
				{
					src: 'https://cdnjs.cloudflare.com/ajax/libs/dragula/3.7.3/dragula.min.css',
					integrity: 'sha512-zlYhSecphd+kwRzeCOyj7/u3HZIQ3Q0NP7AN7ZEKhYTdi0AQOGGbc7eA3I/mUffqjdr8G1/9xoS478h+I0MQGg==',
				},
				{
					src: 'https://cdnjs.cloudflare.com/ajax/libs/dragula/3.7.3/dragula.min.js',
					integrity: 'sha512-NgXVRE+Mxxf647SqmbB9wPS5SEpWiLFp5G7ItUNFi+GVUyQeP+7w4vnKtc2O/Dm74TpTFKXNjakd40pfSKNulg==',
				},
			]).then(function() {
				console.debug('dragula.js loaded');
				// Resolve library reference
				return window.dragula;
			}).fail(function() {
				alertError('Failed to load dragula.js drag and drop library');
			});
	}
}
