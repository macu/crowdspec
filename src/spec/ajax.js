import $ from 'jquery';
import {alertError} from '../utils.js';

export function ajaxLoadSpec(specId, loadBlocks = true) {
	return $.get('/ajax/spec', {
		specId,
		loadBlocks,
	}).fail(alertError);
}

export function ajaxCreateSpec(name, desc, isPublic) {
	return $.post('/ajax/spec/create-spec', {
		name, // must be provided
		desc,
		isPublic,
	}).fail(alertError);
}

export function ajaxSaveSpec(specId, name, desc, isPublic) {
	return $.post('/ajax/spec/save-spec', {
		specId,
		name, // must be provided
		desc,
		isPublic,
	}).fail(alertError);
}

export function ajaxDeleteSpec(specId) {
	return $.post('/ajax/spec/delete-spec', {
		specId,
	}).fail(alertError);
}

export function ajaxLoadSubspecs(specId) {
	return $.get('/ajax/spec/subspecs', {
		specId,
	}).fail(alertError);
}

export function ajaxLoadSubspec(specId, subspecId, loadBlocks = true) {
	return $.get('/ajax/spec/subspec', {
		specId,
		subspecId,
		loadBlocks,
	}).fail(alertError);
}

export function ajaxCreateSubspec(specId, name, desc) {
	return $.post('/ajax/spec/create-subspec', {
		specId, // required
		name, // required
		desc,
	}).fail(alertError);
}

export function ajaxSaveSubspec(subspecId, name, desc) {
	return $.post('/ajax/spec/save-subspec', {
		subspecId, // required
		name, // required
		desc,
	}).fail(alertError);
}

export function ajaxDeleteSubspec(subspecId) {
	return $.post('/ajax/spec/delete-subspec', {
		subspecId,
	}).fail(alertError);
}

export function ajaxCreateBlock(specId, subspecId, parentId, insertBeforeId,
		styleType, contentType, title, body, refType, refFields) {
	refFields = refFields || {};
	return $.post('/ajax/spec/create-block', {
		specId, // must be provided
		subspecId, // null if spec-level
		parentId, // null if no parent
		insertBeforeId, // null for append
		styleType, // must be provided
		contentType, // may be null
		title, // may be null or empty string
		body, // may be null or empty string
		refType, // may be null
		...refFields,
	}).fail(alertError);
}

export function ajaxSaveBlock(specId, blockId,
		styleType, contentType, title, body, refType, refFields) {
	refFields = refFields || {};
	return $.post('/ajax/spec/save-block', {
		specId, // must be provided
		blockId, // must be provided
		styleType, // must be provided
		contentType, // may be null
		title, // may be null or empty string
		body, // may be null or empty string
		refType, // may be null
		...refFields,
	}).fail(alertError);
}

export function ajaxMoveBlocks(blockIds, subspecId, parentId, insertBeforeId) {
	// Go app expects blockIds as an array of ints encoded in a string
	for (var i = 0; i < blockIds.length; i++) {
		blockIds[i] = parseInt(blockIds[i], 10);
	}
	blockIds = JSON.stringify(blockIds);
	return $.post('/ajax/spec/move-blocks', {
		blockIds,
		subspecId, // null if spec-level
		parentId, // null if no parent
		insertBeforeId, // null to insert at end
	}).fail(alertError);
}

export function ajaxDeleteBlock(blockId) {
	return $.post('/ajax/spec/delete-block', {
		blockId,
	}).fail(alertError);
}

export function ajaxLoadUrls(specId) {
	return $.get('/ajax/spec/urls', {
		specId,
	}).fail(alertError);
}

export function ajaxCreateUrl(specId, url) {
	return $.post('/ajax/spec/create-url', {
		specId,
		url,
	}).fail(alertError);
}

export function ajaxRefreshUrl(id, url) {
	return $.post('/ajax/spec/refresh-url', {
		id,
		url,
	}).fail(alertError);
}

export function ajaxDeleteUrl(id) {
	return $.post('/ajax/spec/delete-url', {
		id,
	}).fail(alertError);
}

export function ajaxLoadCommunity(specId, targetType, targetId) {
	return $.get('/ajax/spec/community', {
		specId,
		targetType,
		targetId,
	}).fail(alertError);
}

export function ajaxLoadCommentsPage(specId, targetType, targetId, updatedBefore) {
	// TODO accept filters
	return $.get('/ajax/spec/community/page', {
		specId,
		targetType,
		targetId,
		updatedBefore,
	}).fail(alertError);
}

export function ajaxMarkRead(specId, targetType, targetId, read) {
	return $.post('/ajax/spec/community/mark-read', {
		specId,
		targetType,
		targetId,
		read,
	}).fail(alertError);
}

export function ajaxAddComment(specId, targetType, targetId, body) {
	return $.post('/ajax/spec/community/add-comment', {
		specId,
		targetType,
		targetId,
		body,
	}).fail(alertError);
}

export function ajaxUpdateComment(specId, commentId, body) {
	return $.post('/ajax/spec/community/update-comment', {
		specId,
		commentId,
		body,
	}).fail(alertError);
}

export function ajaxDeleteComment(specId, commentId) {
	return $.post('/ajax/spec/community/delete-comment', {
		specId,
		commentId,
	}).fail(alertError);
}
