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
		name, // must be provided
		desc,
	}).fail(alertError);
}

export function ajaxSaveSpec(specId, name, desc) {
	return $.post('/ajax/spec/save-spec', {
		specId,
		name, // must be provided
		desc,
	}).fail(alertError);
}

export function ajaxDeleteSpec(specId) {
	return $.post('/ajax/spec/delete-spec', {
		specId,
	}).fail(alertError);
}

export function ajaxCreateBlock(specId, subspaceId, parentId, insertBeforeId,
		styleType, contentType, title, body, refFields) {
	refFields = refFields || {};
	return $.post('/ajax/spec/create-block', {
		specId, // must be provided
		subspaceId, // null if spec-level
		parentId, // null if no parent
		insertBeforeId, // null for append
		styleType, // must be provided
		contentType, // may be null
		title, // may be null or empty string
		body, // may be null or empty string
		...refFields,
	}).fail(alertError);
}

export function ajaxSaveBlock(specId, blockId,
		styleType, contentType, title, body, refFields) {
	refFields = refFields || {};
	return $.post('/ajax/spec/save-block', {
		specId, // must be provided
		blockId, // must be provided
		styleType, // must be provided
		contentType, // may be null
		title, // may be null or empty string
		body, // may be null or empty string
		...refFields,
	}).fail(alertError);
}

export function ajaxDeleteBlock(blockId) {
	return $.post('/ajax/spec/delete-block', {
		blockId,
	}).fail(alertError);
}

export function ajaxMoveBlock(blockId, subspaceId, parentId, insertBeforeId) {
	return $.post('/ajax/spec/move-block', {
		blockId, // must be provided
		subspaceId, // null if spec-level
		parentId, // null if no parent
		insertBeforeId, // null to insert at end
	}).fail(alertError);
}

export function fetchUrl(url) {
	return $.get('/ajax/fetch-url', {url}); // No alert on error
}
