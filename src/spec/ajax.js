import $ from 'jquery';
import {alertError} from '../utils.js';

export function ajaxLoadSpec(specId) {
	return $.get('/ajax/spec', {specId}).fail(alertError);
}

export function ajaxCreateSpec(name, desc) {
	return $.post('/ajax/spec/create', {
		name, desc,
	}).fail(alertError);
}

export function ajaxCreateSubpoint(specId, parentId, orderNumber, title, desc) {
	return $.post('/ajax/spec/add-subpoint', {
		specId, parentId, title, desc,
	}).fail(alertError);
}
