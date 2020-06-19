import $ from 'jquery';
import {alertError} from '../utils.js';
import {END_INDEX} from './const.js';

export function ajaxLoadSpec(specId) {
	return $.get('/ajax/spec', {
		specId,
	}).fail(alertError);
}

export function ajaxCreateSpec(name, desc) {
	return $.post('/ajax/spec/create-spec', {
		name,
		desc,
	}).fail(alertError);
}

export function ajaxCreateBlock(specId, subspaceId, parentId, insertAt,
		styleType, contentType, refType, refId, title, body) {
	return $.post('/ajax/spec/create-block', {
		specId, // must be provided
		subspaceId, // null if spec-level
		parentId, // null if no parent
		insertAt: (insertAt || insertAt === 0) ? insertAt : END_INDEX, // default insert at end
		styleType, // must be provided
		contentType, // may be null
		refType, // may be null
		refId, // required if refType given
		title, // may be null or empty string
		body, // may be null or empty string
	}).fail(alertError);
}
