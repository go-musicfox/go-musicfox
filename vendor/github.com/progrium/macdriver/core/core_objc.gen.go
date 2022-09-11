//go:build darwin
// +build darwin

package core

import (
	"github.com/progrium/macdriver/objc"
	"unsafe"
)

/*
#cgo CFLAGS: -x objective-c -Wno-everything
#cgo LDFLAGS: -lobjc -framework AppKit -framework QuartzCore -framework Foundation
#define __OBJC2__ 1
#include <objc/message.h>
#include <stdlib.h>

#include <AppKit/AppKit.h>
#include <QuartzCore/QuartzCore.h>
#include <Foundation/Foundation.h>

bool core_convertObjCBool(BOOL b) {
	if (b) { return true; }
	return false;
}


void* CALayer_type_alloc() {
	return [CALayer
		alloc];
}
void* CALayer_type_layer() {
	return [CALayer
		layer];
}
void* CALayer_type_layerWithRemoteClientId_(uint32_t client_id) {
	return [CALayer
		layerWithRemoteClientId: client_id];
}
BOOL CALayer_type_needsDisplayForKey_(void* key) {
	return [CALayer
		needsDisplayForKey: key];
}
void* CALayer_type_defaultActionForKey_(void* event) {
	return [CALayer
		defaultActionForKey: event];
}
void* CALayer_type_defaultValueForKey_(void* key) {
	return [CALayer
		defaultValueForKey: key];
}
void* NSArray_type_alloc() {
	return [NSArray
		alloc];
}
void* NSArray_type_array() {
	return [NSArray
		array];
}
void* NSArray_type_arrayWithArray_(void* array) {
	return [NSArray
		arrayWithArray: array];
}
void* NSArray_type_arrayWithObject_(void* anObject) {
	return [NSArray
		arrayWithObject: anObject];
}
void* NSArray_type_arrayWithContentsOfURL_error_(void* url, void* error) {
	return [NSArray
		arrayWithContentsOfURL: url
		error: error];
}
void* NSAttributedString_type_alloc() {
	return [NSAttributedString
		alloc];
}
void* NSAttributedString_type_textTypes() {
	return [NSAttributedString
		textTypes];
}
void* NSAttributedString_type_textUnfilteredTypes() {
	return [NSAttributedString
		textUnfilteredTypes];
}
void* NSData_type_alloc() {
	return [NSData
		alloc];
}
void* NSData_type_data() {
	return [NSData
		data];
}
void* NSData_type_dataWithBytes_length_(void* bytes, unsigned long length) {
	return [NSData
		dataWithBytes: bytes
		length: length];
}
void* NSData_type_dataWithBytesNoCopy_length_(void* bytes, unsigned long length) {
	return [NSData
		dataWithBytesNoCopy: bytes
		length: length];
}
void* NSData_type_dataWithBytesNoCopy_length_freeWhenDone_(void* bytes, unsigned long length, BOOL b) {
	return [NSData
		dataWithBytesNoCopy: bytes
		length: length
		freeWhenDone: b];
}
void* NSData_type_dataWithData_(void* data) {
	return [NSData
		dataWithData: data];
}
void* NSData_type_dataWithContentsOfFile_(void* path) {
	return [NSData
		dataWithContentsOfFile: path];
}
void* NSData_type_dataWithContentsOfURL_(void* url) {
	return [NSData
		dataWithContentsOfURL: url];
}
void* NSDate_type_alloc() {
	return [NSDate
		alloc];
}
void* NSDate_type_date() {
	return [NSDate
		date];
}
void* NSDate_type_distantFuture() {
	return [NSDate
		distantFuture];
}
void* NSDate_type_distantPast() {
	return [NSDate
		distantPast];
}
void* NSDate_type_now() {
	return [NSDate
		now];
}
void* NSDictionary_type_alloc() {
	return [NSDictionary
		alloc];
}
void* NSDictionary_type_dictionary() {
	return [NSDictionary
		dictionary];
}
void* NSDictionary_type_dictionaryWithObjects_forKeys_(void* objects, void* keys) {
	return [NSDictionary
		dictionaryWithObjects: objects
		forKeys: keys];
}
void* NSDictionary_type_dictionaryWithObject_forKey_(void* object, void* key) {
	return [NSDictionary
		dictionaryWithObject: object
		forKey: key];
}
void* NSDictionary_type_dictionaryWithDictionary_(void* dict) {
	return [NSDictionary
		dictionaryWithDictionary: dict];
}
void* NSDictionary_type_dictionaryWithContentsOfURL_error_(void* url, void* error) {
	return [NSDictionary
		dictionaryWithContentsOfURL: url
		error: error];
}
void* NSDictionary_type_sharedKeySetForKeys_(void* keys) {
	return [NSDictionary
		sharedKeySetForKeys: keys];
}
void* NSNumber_type_alloc() {
	return [NSNumber
		alloc];
}
void* NSNumber_type_numberWithBool_(BOOL value) {
	return [NSNumber
		numberWithBool: value];
}
void* NSNumber_type_numberWithDouble_(double value) {
	return [NSNumber
		numberWithDouble: value];
}
void* NSNumber_type_numberWithFloat_(float value) {
	return [NSNumber
		numberWithFloat: value];
}
void* NSNumber_type_numberWithInt_(int value) {
	return [NSNumber
		numberWithInt: value];
}
void* NSNumber_type_numberWithInteger_(long value) {
	return [NSNumber
		numberWithInteger: value];
}
void* NSNumber_type_numberWithUnsignedInt_(int value) {
	return [NSNumber
		numberWithUnsignedInt: value];
}
void* NSNumber_type_numberWithUnsignedInteger_(unsigned long value) {
	return [NSNumber
		numberWithUnsignedInteger: value];
}
void* NSRunLoop_type_alloc() {
	return [NSRunLoop
		alloc];
}
void* NSRunLoop_type_currentRunLoop() {
	return [NSRunLoop
		currentRunLoop];
}
void* NSRunLoop_type_mainRunLoop() {
	return [NSRunLoop
		mainRunLoop];
}
void* NSString_type_alloc() {
	return [NSString
		alloc];
}
void* NSString_type_string() {
	return [NSString
		string];
}
void* NSString_type_localizedUserNotificationStringForKey_arguments_(void* key, void* arguments) {
	return [NSString
		localizedUserNotificationStringForKey: key
		arguments: arguments];
}
void* NSString_type_stringWithString_(void* string) {
	return [NSString
		stringWithString: string];
}
void* NSString_type_stringWithContentsOfFile_encoding_error_(void* path, unsigned long enc, void* error) {
	return [NSString
		stringWithContentsOfFile: path
		encoding: enc
		error: error];
}
void* NSString_type_stringWithContentsOfURL_encoding_error_(void* url, unsigned long enc, void* error) {
	return [NSString
		stringWithContentsOfURL: url
		encoding: enc
		error: error];
}
void* NSString_type_localizedNameOfStringEncoding_(unsigned long encoding) {
	return [NSString
		localizedNameOfStringEncoding: encoding];
}
void* NSString_type_pathWithComponents_(void* components) {
	return [NSString
		pathWithComponents: components];
}
unsigned long NSString_type_defaultCStringEncoding() {
	return [NSString
		defaultCStringEncoding];
}
void* NSError_type_alloc() {
	return [NSError
		alloc];
}
void* NSError_type_errorWithDomain_code_userInfo_(void* domain, long code, void* dict) {
	return [NSError
		errorWithDomain: domain
		code: code
		userInfo: dict];
}
void* NSThread_type_alloc() {
	return [NSThread
		alloc];
}
void NSThread_type_detachNewThreadSelector_toTarget_withObject_(void* selector, void* target, void* argument) {
	[NSThread
		detachNewThreadSelector: selector
		toTarget: target
		withObject: argument];
}
void NSThread_type_sleepUntilDate_(void* date) {
	[NSThread
		sleepUntilDate: date];
}
void NSThread_type_sleepForTimeInterval_(NSTimeInterval ti) {
	[NSThread
		sleepForTimeInterval: ti];
}
void NSThread_type_exit() {
	[NSThread
		exit];
}
BOOL NSThread_type_isMultiThreaded() {
	return [NSThread
		isMultiThreaded];
}
double NSThread_type_threadPriority() {
	return [NSThread
		threadPriority];
}
BOOL NSThread_type_setThreadPriority_(double p) {
	return [NSThread
		setThreadPriority: p];
}
BOOL NSThread_type_isMainThread() {
	return [NSThread
		isMainThread];
}
void* NSThread_type_mainThread() {
	return [NSThread
		mainThread];
}
void* NSThread_type_currentThread() {
	return [NSThread
		currentThread];
}
void* NSThread_type_callStackReturnAddresses() {
	return [NSThread
		callStackReturnAddresses];
}
void* NSThread_type_callStackSymbols() {
	return [NSThread
		callStackSymbols];
}
void* NSURL_type_alloc() {
	return [NSURL
		alloc];
}
void* NSURL_type_URLWithString_(void* URLString) {
	return [NSURL
		URLWithString: URLString];
}
void* NSURL_type_URLWithString_relativeToURL_(void* URLString, void* baseURL) {
	return [NSURL
		URLWithString: URLString
		relativeToURL: baseURL];
}
void* NSURL_type_fileURLWithPath_isDirectory_(void* path, BOOL isDir) {
	return [NSURL
		fileURLWithPath: path
		isDirectory: isDir];
}
void* NSURL_type_fileURLWithPath_relativeToURL_(void* path, void* baseURL) {
	return [NSURL
		fileURLWithPath: path
		relativeToURL: baseURL];
}
void* NSURL_type_fileURLWithPath_isDirectory_relativeToURL_(void* path, BOOL isDir, void* baseURL) {
	return [NSURL
		fileURLWithPath: path
		isDirectory: isDir
		relativeToURL: baseURL];
}
void* NSURL_type_fileURLWithPath_(void* path) {
	return [NSURL
		fileURLWithPath: path];
}
void* NSURL_type_fileURLWithPathComponents_(void* components) {
	return [NSURL
		fileURLWithPathComponents: components];
}
void* NSURL_type_absoluteURLWithDataRepresentation_relativeToURL_(void* data, void* baseURL) {
	return [NSURL
		absoluteURLWithDataRepresentation: data
		relativeToURL: baseURL];
}
void* NSURL_type_URLWithDataRepresentation_relativeToURL_(void* data, void* baseURL) {
	return [NSURL
		URLWithDataRepresentation: data
		relativeToURL: baseURL];
}
void* NSURL_type_bookmarkDataWithContentsOfURL_error_(void* bookmarkFileURL, void* error) {
	return [NSURL
		bookmarkDataWithContentsOfURL: bookmarkFileURL
		error: error];
}
void* NSURL_type_resourceValuesForKeys_fromBookmarkData_(void* keys, void* bookmarkData) {
	return [NSURL
		resourceValuesForKeys: keys
		fromBookmarkData: bookmarkData];
}
void* NSURLRequest_type_alloc() {
	return [NSURLRequest
		alloc];
}
void* NSURLRequest_type_requestWithURL_(void* URL) {
	return [NSURLRequest
		requestWithURL: URL];
}
BOOL NSURLRequest_type_supportsSecureCoding() {
	return [NSURLRequest
		supportsSecureCoding];
}
void* NSNotification_type_alloc() {
	return [NSNotification
		alloc];
}
void* NSOperationQueue_type_alloc() {
	return [NSOperationQueue
		alloc];
}
void* NSOperationQueue_type_mainQueue() {
	return [NSOperationQueue
		mainQueue];
}
void* NSOperationQueue_type_currentQueue() {
	return [NSOperationQueue
		currentQueue];
}
void* NSNotificationCenter_type_alloc() {
	return [NSNotificationCenter
		alloc];
}
void* NSNotificationCenter_type_defaultCenter() {
	return [NSNotificationCenter
		defaultCenter];
}


void* CALayer_inst_init(void *id) {
	return [(CALayer*)id
		init];
}

void* CALayer_inst_initWithLayer_(void *id, void* layer) {
	return [(CALayer*)id
		initWithLayer: layer];
}

void* CALayer_inst_presentationLayer(void *id) {
	return [(CALayer*)id
		presentationLayer];
}

void* CALayer_inst_modelLayer(void *id) {
	return [(CALayer*)id
		modelLayer];
}

void CALayer_inst_display(void *id) {
	[(CALayer*)id
		display];
}

BOOL CALayer_inst_contentsAreFlipped(void *id) {
	return [(CALayer*)id
		contentsAreFlipped];
}

void CALayer_inst_addSublayer_(void *id, void* layer) {
	[(CALayer*)id
		addSublayer: layer];
}

void CALayer_inst_removeFromSuperlayer(void *id) {
	[(CALayer*)id
		removeFromSuperlayer];
}

void CALayer_inst_insertSublayer_atIndex_(void *id, void* layer, int idx) {
	[(CALayer*)id
		insertSublayer: layer
		atIndex: idx];
}

void CALayer_inst_insertSublayer_below_(void *id, void* layer, void* sibling) {
	[(CALayer*)id
		insertSublayer: layer
		below: sibling];
}

void CALayer_inst_insertSublayer_above_(void *id, void* layer, void* sibling) {
	[(CALayer*)id
		insertSublayer: layer
		above: sibling];
}

void CALayer_inst_replaceSublayer_with_(void *id, void* oldLayer, void* newLayer) {
	[(CALayer*)id
		replaceSublayer: oldLayer
		with: newLayer];
}

void CALayer_inst_setNeedsDisplay(void *id) {
	[(CALayer*)id
		setNeedsDisplay];
}

void CALayer_inst_setNeedsDisplayInRect_(void *id, NSRect r) {
	[(CALayer*)id
		setNeedsDisplayInRect: r];
}

void CALayer_inst_displayIfNeeded(void *id) {
	[(CALayer*)id
		displayIfNeeded];
}

BOOL CALayer_inst_needsDisplay(void *id) {
	return [(CALayer*)id
		needsDisplay];
}

void CALayer_inst_removeAllAnimations(void *id) {
	[(CALayer*)id
		removeAllAnimations];
}

void CALayer_inst_removeAnimationForKey_(void *id, void* key) {
	[(CALayer*)id
		removeAnimationForKey: key];
}

void* CALayer_inst_animationKeys(void *id) {
	return [(CALayer*)id
		animationKeys];
}

void CALayer_inst_setNeedsLayout(void *id) {
	[(CALayer*)id
		setNeedsLayout];
}

void CALayer_inst_layoutSublayers(void *id) {
	[(CALayer*)id
		layoutSublayers];
}

void CALayer_inst_layoutIfNeeded(void *id) {
	[(CALayer*)id
		layoutIfNeeded];
}

BOOL CALayer_inst_needsLayout(void *id) {
	return [(CALayer*)id
		needsLayout];
}

void CALayer_inst_resizeWithOldSuperlayerSize_(void *id, NSSize size) {
	[(CALayer*)id
		resizeWithOldSuperlayerSize: size];
}

void CALayer_inst_resizeSublayersWithOldSize_(void *id, NSSize size) {
	[(CALayer*)id
		resizeSublayersWithOldSize: size];
}

NSSize CALayer_inst_preferredFrameSize(void *id) {
	return [(CALayer*)id
		preferredFrameSize];
}

void* CALayer_inst_actionForKey_(void *id, void* event) {
	return [(CALayer*)id
		actionForKey: event];
}

NSRect CALayer_inst_convertRect_fromLayer_(void *id, NSRect r, void* l) {
	return [(CALayer*)id
		convertRect: r
		fromLayer: l];
}

NSRect CALayer_inst_convertRect_toLayer_(void *id, NSRect r, void* l) {
	return [(CALayer*)id
		convertRect: r
		toLayer: l];
}

void CALayer_inst_scrollRectToVisible_(void *id, NSRect r) {
	[(CALayer*)id
		scrollRectToVisible: r];
}

BOOL CALayer_inst_shouldArchiveValueForKey_(void *id, void* key) {
	return [(CALayer*)id
		shouldArchiveValueForKey: key];
}

void* CALayer_inst_delegate(void *id) {
	return [(CALayer*)id
		delegate];
}

void CALayer_inst_setDelegate_(void *id, void* value) {
	[(CALayer*)id
		setDelegate: value];
}

void* CALayer_inst_contents(void *id) {
	return [(CALayer*)id
		contents];
}

void CALayer_inst_setContents_(void *id, void* value) {
	[(CALayer*)id
		setContents: value];
}

NSRect CALayer_inst_contentsRect(void *id) {
	return [(CALayer*)id
		contentsRect];
}

void CALayer_inst_setContentsRect_(void *id, NSRect value) {
	[(CALayer*)id
		setContentsRect: value];
}

NSRect CALayer_inst_contentsCenter(void *id) {
	return [(CALayer*)id
		contentsCenter];
}

void CALayer_inst_setContentsCenter_(void *id, NSRect value) {
	[(CALayer*)id
		setContentsCenter: value];
}

float CALayer_inst_opacity(void *id) {
	return [(CALayer*)id
		opacity];
}

void CALayer_inst_setOpacity_(void *id, float value) {
	[(CALayer*)id
		setOpacity: value];
}

BOOL CALayer_inst_isHidden(void *id) {
	return [(CALayer*)id
		isHidden];
}

void CALayer_inst_setHidden_(void *id, BOOL value) {
	[(CALayer*)id
		setHidden: value];
}

BOOL CALayer_inst_masksToBounds(void *id) {
	return [(CALayer*)id
		masksToBounds];
}

void CALayer_inst_setMasksToBounds_(void *id, BOOL value) {
	[(CALayer*)id
		setMasksToBounds: value];
}

void* CALayer_inst_mask(void *id) {
	return [(CALayer*)id
		mask];
}

void CALayer_inst_setMask_(void *id, void* value) {
	[(CALayer*)id
		setMask: value];
}

BOOL CALayer_inst_isDoubleSided(void *id) {
	return [(CALayer*)id
		isDoubleSided];
}

void CALayer_inst_setDoubleSided_(void *id, BOOL value) {
	[(CALayer*)id
		setDoubleSided: value];
}

double CALayer_inst_cornerRadius(void *id) {
	return [(CALayer*)id
		cornerRadius];
}

void CALayer_inst_setCornerRadius_(void *id, double value) {
	[(CALayer*)id
		setCornerRadius: value];
}

double CALayer_inst_borderWidth(void *id) {
	return [(CALayer*)id
		borderWidth];
}

void CALayer_inst_setBorderWidth_(void *id, double value) {
	[(CALayer*)id
		setBorderWidth: value];
}

float CALayer_inst_shadowOpacity(void *id) {
	return [(CALayer*)id
		shadowOpacity];
}

void CALayer_inst_setShadowOpacity_(void *id, float value) {
	[(CALayer*)id
		setShadowOpacity: value];
}

double CALayer_inst_shadowRadius(void *id) {
	return [(CALayer*)id
		shadowRadius];
}

void CALayer_inst_setShadowRadius_(void *id, double value) {
	[(CALayer*)id
		setShadowRadius: value];
}

NSSize CALayer_inst_shadowOffset(void *id) {
	return [(CALayer*)id
		shadowOffset];
}

void CALayer_inst_setShadowOffset_(void *id, NSSize value) {
	[(CALayer*)id
		setShadowOffset: value];
}

void* CALayer_inst_style(void *id) {
	return [(CALayer*)id
		style];
}

void CALayer_inst_setStyle_(void *id, void* value) {
	[(CALayer*)id
		setStyle: value];
}

BOOL CALayer_inst_allowsEdgeAntialiasing(void *id) {
	return [(CALayer*)id
		allowsEdgeAntialiasing];
}

void CALayer_inst_setAllowsEdgeAntialiasing_(void *id, BOOL value) {
	[(CALayer*)id
		setAllowsEdgeAntialiasing: value];
}

BOOL CALayer_inst_allowsGroupOpacity(void *id) {
	return [(CALayer*)id
		allowsGroupOpacity];
}

void CALayer_inst_setAllowsGroupOpacity_(void *id, BOOL value) {
	[(CALayer*)id
		setAllowsGroupOpacity: value];
}

void* CALayer_inst_filters(void *id) {
	return [(CALayer*)id
		filters];
}

void CALayer_inst_setFilters_(void *id, void* value) {
	[(CALayer*)id
		setFilters: value];
}

void* CALayer_inst_compositingFilter(void *id) {
	return [(CALayer*)id
		compositingFilter];
}

void CALayer_inst_setCompositingFilter_(void *id, void* value) {
	[(CALayer*)id
		setCompositingFilter: value];
}

void* CALayer_inst_backgroundFilters(void *id) {
	return [(CALayer*)id
		backgroundFilters];
}

void CALayer_inst_setBackgroundFilters_(void *id, void* value) {
	[(CALayer*)id
		setBackgroundFilters: value];
}

float CALayer_inst_minificationFilterBias(void *id) {
	return [(CALayer*)id
		minificationFilterBias];
}

void CALayer_inst_setMinificationFilterBias_(void *id, float value) {
	[(CALayer*)id
		setMinificationFilterBias: value];
}

BOOL CALayer_inst_isOpaque(void *id) {
	return [(CALayer*)id
		isOpaque];
}

void CALayer_inst_setOpaque_(void *id, BOOL value) {
	[(CALayer*)id
		setOpaque: value];
}

BOOL CALayer_inst_isGeometryFlipped(void *id) {
	return [(CALayer*)id
		isGeometryFlipped];
}

void CALayer_inst_setGeometryFlipped_(void *id, BOOL value) {
	[(CALayer*)id
		setGeometryFlipped: value];
}

BOOL CALayer_inst_drawsAsynchronously(void *id) {
	return [(CALayer*)id
		drawsAsynchronously];
}

void CALayer_inst_setDrawsAsynchronously_(void *id, BOOL value) {
	[(CALayer*)id
		setDrawsAsynchronously: value];
}

BOOL CALayer_inst_shouldRasterize(void *id) {
	return [(CALayer*)id
		shouldRasterize];
}

void CALayer_inst_setShouldRasterize_(void *id, BOOL value) {
	[(CALayer*)id
		setShouldRasterize: value];
}

double CALayer_inst_rasterizationScale(void *id) {
	return [(CALayer*)id
		rasterizationScale];
}

void CALayer_inst_setRasterizationScale_(void *id, double value) {
	[(CALayer*)id
		setRasterizationScale: value];
}

NSRect CALayer_inst_frame(void *id) {
	return [(CALayer*)id
		frame];
}

void CALayer_inst_setFrame_(void *id, NSRect value) {
	[(CALayer*)id
		setFrame: value];
}

NSRect CALayer_inst_bounds(void *id) {
	return [(CALayer*)id
		bounds];
}

void CALayer_inst_setBounds_(void *id, NSRect value) {
	[(CALayer*)id
		setBounds: value];
}

double CALayer_inst_zPosition(void *id) {
	return [(CALayer*)id
		zPosition];
}

void CALayer_inst_setZPosition_(void *id, double value) {
	[(CALayer*)id
		setZPosition: value];
}

double CALayer_inst_anchorPointZ(void *id) {
	return [(CALayer*)id
		anchorPointZ];
}

void CALayer_inst_setAnchorPointZ_(void *id, double value) {
	[(CALayer*)id
		setAnchorPointZ: value];
}

double CALayer_inst_contentsScale(void *id) {
	return [(CALayer*)id
		contentsScale];
}

void CALayer_inst_setContentsScale_(void *id, double value) {
	[(CALayer*)id
		setContentsScale: value];
}

void* CALayer_inst_sublayers(void *id) {
	return [(CALayer*)id
		sublayers];
}

void CALayer_inst_setSublayers_(void *id, void* value) {
	[(CALayer*)id
		setSublayers: value];
}

void* CALayer_inst_superlayer(void *id) {
	return [(CALayer*)id
		superlayer];
}

BOOL CALayer_inst_needsDisplayOnBoundsChange(void *id) {
	return [(CALayer*)id
		needsDisplayOnBoundsChange];
}

void CALayer_inst_setNeedsDisplayOnBoundsChange_(void *id, BOOL value) {
	[(CALayer*)id
		setNeedsDisplayOnBoundsChange: value];
}

void* CALayer_inst_layoutManager(void *id) {
	return [(CALayer*)id
		layoutManager];
}

void CALayer_inst_setLayoutManager_(void *id, void* value) {
	[(CALayer*)id
		setLayoutManager: value];
}

void* CALayer_inst_constraints(void *id) {
	return [(CALayer*)id
		constraints];
}

void CALayer_inst_setConstraints_(void *id, void* value) {
	[(CALayer*)id
		setConstraints: value];
}

void* CALayer_inst_actions(void *id) {
	return [(CALayer*)id
		actions];
}

void CALayer_inst_setActions_(void *id, void* value) {
	[(CALayer*)id
		setActions: value];
}

NSRect CALayer_inst_visibleRect(void *id) {
	return [(CALayer*)id
		visibleRect];
}

void* CALayer_inst_name(void *id) {
	return [(CALayer*)id
		name];
}

void CALayer_inst_setName_(void *id, void* value) {
	[(CALayer*)id
		setName: value];
}

void* NSArray_inst_init(void *id) {
	return [(NSArray*)id
		init];
}

void* NSArray_inst_initWithArray_(void *id, void* array) {
	return [(NSArray*)id
		initWithArray: array];
}

void* NSArray_inst_initWithArray_copyItems_(void *id, void* array, BOOL flag) {
	return [(NSArray*)id
		initWithArray: array
		copyItems: flag];
}

BOOL NSArray_inst_containsObject_(void *id, void* anObject) {
	return [(NSArray*)id
		containsObject: anObject];
}

void* NSArray_inst_objectAtIndex_(void *id, unsigned long index) {
	return [(NSArray*)id
		objectAtIndex: index];
}

void* NSArray_inst_objectAtIndexedSubscript_(void *id, unsigned long idx) {
	return [(NSArray*)id
		objectAtIndexedSubscript: idx];
}

unsigned long NSArray_inst_indexOfObject_(void *id, void* anObject) {
	return [(NSArray*)id
		indexOfObject: anObject];
}

unsigned long NSArray_inst_indexOfObjectIdenticalTo_(void *id, void* anObject) {
	return [(NSArray*)id
		indexOfObjectIdenticalTo: anObject];
}

void NSArray_inst_makeObjectsPerformSelector_(void *id, void* aSelector) {
	[(NSArray*)id
		makeObjectsPerformSelector: aSelector];
}

void NSArray_inst_makeObjectsPerformSelector_withObject_(void *id, void* aSelector, void* argument) {
	[(NSArray*)id
		makeObjectsPerformSelector: aSelector
		withObject: argument];
}

void* NSArray_inst_firstObjectCommonWithArray_(void *id, void* otherArray) {
	return [(NSArray*)id
		firstObjectCommonWithArray: otherArray];
}

BOOL NSArray_inst_isEqualToArray_(void *id, void* otherArray) {
	return [(NSArray*)id
		isEqualToArray: otherArray];
}

void* NSArray_inst_arrayByAddingObject_(void *id, void* anObject) {
	return [(NSArray*)id
		arrayByAddingObject: anObject];
}

void* NSArray_inst_arrayByAddingObjectsFromArray_(void *id, void* otherArray) {
	return [(NSArray*)id
		arrayByAddingObjectsFromArray: otherArray];
}

void* NSArray_inst_sortedArrayUsingDescriptors_(void *id, void* sortDescriptors) {
	return [(NSArray*)id
		sortedArrayUsingDescriptors: sortDescriptors];
}

void* NSArray_inst_sortedArrayUsingSelector_(void *id, void* comparator) {
	return [(NSArray*)id
		sortedArrayUsingSelector: comparator];
}

void* NSArray_inst_componentsJoinedByString_(void *id, void* separator) {
	return [(NSArray*)id
		componentsJoinedByString: separator];
}

void* NSArray_inst_descriptionWithLocale_(void *id, void* locale) {
	return [(NSArray*)id
		descriptionWithLocale: locale];
}

void* NSArray_inst_descriptionWithLocale_indent_(void *id, void* locale, unsigned long level) {
	return [(NSArray*)id
		descriptionWithLocale: locale
		indent: level];
}

void* NSArray_inst_pathsMatchingExtensions_(void *id, void* filterTypes) {
	return [(NSArray*)id
		pathsMatchingExtensions: filterTypes];
}

void NSArray_inst_setValue_forKey_(void *id, void* value, void* key) {
	[(NSArray*)id
		setValue: value
		forKey: key];
}

void* NSArray_inst_valueForKey_(void *id, void* key) {
	return [(NSArray*)id
		valueForKey: key];
}

void* NSArray_inst_shuffledArray(void *id) {
	return [(NSArray*)id
		shuffledArray];
}

void* NSArray_inst_initWithContentsOfURL_error_(void *id, void* url, void* error) {
	return [(NSArray*)id
		initWithContentsOfURL: url
		error: error];
}

BOOL NSArray_inst_writeToURL_error_(void *id, void* url, void* error) {
	return [(NSArray*)id
		writeToURL: url
		error: error];
}

unsigned long NSArray_inst_count(void *id) {
	return [(NSArray*)id
		count];
}

void* NSArray_inst_firstObject(void *id) {
	return [(NSArray*)id
		firstObject];
}

void* NSArray_inst_lastObject(void *id) {
	return [(NSArray*)id
		lastObject];
}

void* NSArray_inst_sortedArrayHint(void *id) {
	return [(NSArray*)id
		sortedArrayHint];
}

void* NSArray_inst_description(void *id) {
	return [(NSArray*)id
		description];
}

void* NSAttributedString_inst_initWithString_(void *id, void* str) {
	return [(NSAttributedString*)id
		initWithString: str];
}

void* NSAttributedString_inst_initWithString_attributes_(void *id, void* str, void* attrs) {
	return [(NSAttributedString*)id
		initWithString: str
		attributes: attrs];
}

void* NSAttributedString_inst_initWithAttributedString_(void *id, void* attrStr) {
	return [(NSAttributedString*)id
		initWithAttributedString: attrStr];
}

void* NSAttributedString_inst_initWithData_options_documentAttributes_error_(void *id, void* data, void* options, void* dict, void* error) {
	return [(NSAttributedString*)id
		initWithData: data
		options: options
		documentAttributes: dict
		error: error];
}

void* NSAttributedString_inst_initWithDocFormat_documentAttributes_(void *id, void* data, void* dict) {
	return [(NSAttributedString*)id
		initWithDocFormat: data
		documentAttributes: dict];
}

void* NSAttributedString_inst_initWithHTML_documentAttributes_(void *id, void* data, void* dict) {
	return [(NSAttributedString*)id
		initWithHTML: data
		documentAttributes: dict];
}

void* NSAttributedString_inst_initWithHTML_baseURL_documentAttributes_(void *id, void* data, void* base, void* dict) {
	return [(NSAttributedString*)id
		initWithHTML: data
		baseURL: base
		documentAttributes: dict];
}

void* NSAttributedString_inst_initWithHTML_options_documentAttributes_(void *id, void* data, void* options, void* dict) {
	return [(NSAttributedString*)id
		initWithHTML: data
		options: options
		documentAttributes: dict];
}

void* NSAttributedString_inst_initWithRTF_documentAttributes_(void *id, void* data, void* dict) {
	return [(NSAttributedString*)id
		initWithRTF: data
		documentAttributes: dict];
}

void* NSAttributedString_inst_initWithRTFD_documentAttributes_(void *id, void* data, void* dict) {
	return [(NSAttributedString*)id
		initWithRTFD: data
		documentAttributes: dict];
}

void* NSAttributedString_inst_initWithURL_options_documentAttributes_error_(void *id, void* url, void* options, void* dict, void* error) {
	return [(NSAttributedString*)id
		initWithURL: url
		options: options
		documentAttributes: dict
		error: error];
}

BOOL NSAttributedString_inst_isEqualToAttributedString_(void *id, void* other) {
	return [(NSAttributedString*)id
		isEqualToAttributedString: other];
}

unsigned long NSAttributedString_inst_nextWordFromIndex_forward_(void *id, unsigned long location, BOOL isForward) {
	return [(NSAttributedString*)id
		nextWordFromIndex: location
		forward: isForward];
}

void NSAttributedString_inst_drawInRect_(void *id, NSRect rect) {
	[(NSAttributedString*)id
		drawInRect: rect];
}

NSSize NSAttributedString_inst_size(void *id) {
	return [(NSAttributedString*)id
		size];
}

void* NSAttributedString_inst_init(void *id) {
	return [(NSAttributedString*)id
		init];
}

void* NSAttributedString_inst_string(void *id) {
	return [(NSAttributedString*)id
		string];
}

unsigned long NSAttributedString_inst_length(void *id) {
	return [(NSAttributedString*)id
		length];
}

void* NSData_inst_initWithBytes_length_(void *id, void* bytes, unsigned long length) {
	return [(NSData*)id
		initWithBytes: bytes
		length: length];
}

void* NSData_inst_initWithBytesNoCopy_length_(void *id, void* bytes, unsigned long length) {
	return [(NSData*)id
		initWithBytesNoCopy: bytes
		length: length];
}

void* NSData_inst_initWithBytesNoCopy_length_freeWhenDone_(void *id, void* bytes, unsigned long length, BOOL b) {
	return [(NSData*)id
		initWithBytesNoCopy: bytes
		length: length
		freeWhenDone: b];
}

void* NSData_inst_initWithData_(void *id, void* data) {
	return [(NSData*)id
		initWithData: data];
}

void* NSData_inst_initWithContentsOfFile_(void *id, void* path) {
	return [(NSData*)id
		initWithContentsOfFile: path];
}

void* NSData_inst_initWithContentsOfURL_(void *id, void* url) {
	return [(NSData*)id
		initWithContentsOfURL: url];
}

BOOL NSData_inst_writeToFile_atomically_(void *id, void* path, BOOL useAuxiliaryFile) {
	return [(NSData*)id
		writeToFile: path
		atomically: useAuxiliaryFile];
}

BOOL NSData_inst_writeToURL_atomically_(void *id, void* url, BOOL atomically) {
	return [(NSData*)id
		writeToURL: url
		atomically: atomically];
}

void NSData_inst_getBytes_length_(void *id, void* buffer, unsigned long length) {
	[(NSData*)id
		getBytes: buffer
		length: length];
}

BOOL NSData_inst_isEqualToData_(void *id, void* other) {
	return [(NSData*)id
		isEqualToData: other];
}

void* NSData_inst_init(void *id) {
	return [(NSData*)id
		init];
}

void* NSData_inst_bytes(void *id) {
	return [(NSData*)id
		bytes];
}

unsigned long NSData_inst_length(void *id) {
	return [(NSData*)id
		length];
}

void* NSData_inst_description(void *id) {
	return [(NSData*)id
		description];
}

void* NSDate_inst_init(void *id) {
	return [(NSDate*)id
		init];
}

BOOL NSDate_inst_isEqualToDate_(void *id, void* otherDate) {
	return [(NSDate*)id
		isEqualToDate: otherDate];
}

void* NSDate_inst_earlierDate_(void *id, void* anotherDate) {
	return [(NSDate*)id
		earlierDate: anotherDate];
}

void* NSDate_inst_laterDate_(void *id, void* anotherDate) {
	return [(NSDate*)id
		laterDate: anotherDate];
}

void* NSDate_inst_descriptionWithLocale_(void *id, void* locale) {
	return [(NSDate*)id
		descriptionWithLocale: locale];
}

void* NSDate_inst_description(void *id) {
	return [(NSDate*)id
		description];
}

void* NSDictionary_inst_init(void *id) {
	return [(NSDictionary*)id
		init];
}

void* NSDictionary_inst_initWithObjects_forKeys_(void *id, void* objects, void* keys) {
	return [(NSDictionary*)id
		initWithObjects: objects
		forKeys: keys];
}

void* NSDictionary_inst_initWithDictionary_(void *id, void* otherDictionary) {
	return [(NSDictionary*)id
		initWithDictionary: otherDictionary];
}

void* NSDictionary_inst_initWithDictionary_copyItems_(void *id, void* otherDictionary, BOOL flag) {
	return [(NSDictionary*)id
		initWithDictionary: otherDictionary
		copyItems: flag];
}

void* NSDictionary_inst_initWithContentsOfURL_error_(void *id, void* url, void* error) {
	return [(NSDictionary*)id
		initWithContentsOfURL: url
		error: error];
}

BOOL NSDictionary_inst_isEqualToDictionary_(void *id, void* otherDictionary) {
	return [(NSDictionary*)id
		isEqualToDictionary: otherDictionary];
}

void* NSDictionary_inst_allKeysForObject_(void *id, void* anObject) {
	return [(NSDictionary*)id
		allKeysForObject: anObject];
}

void* NSDictionary_inst_valueForKey_(void *id, void* key) {
	return [(NSDictionary*)id
		valueForKey: key];
}

void* NSDictionary_inst_objectsForKeys_notFoundMarker_(void *id, void* keys, void* marker) {
	return [(NSDictionary*)id
		objectsForKeys: keys
		notFoundMarker: marker];
}

void* NSDictionary_inst_keysSortedByValueUsingSelector_(void *id, void* comparator) {
	return [(NSDictionary*)id
		keysSortedByValueUsingSelector: comparator];
}

BOOL NSDictionary_inst_writeToURL_error_(void *id, void* url, void* error) {
	return [(NSDictionary*)id
		writeToURL: url
		error: error];
}

void* NSDictionary_inst_fileType(void *id) {
	return [(NSDictionary*)id
		fileType];
}

void* NSDictionary_inst_fileCreationDate(void *id) {
	return [(NSDictionary*)id
		fileCreationDate];
}

void* NSDictionary_inst_fileModificationDate(void *id) {
	return [(NSDictionary*)id
		fileModificationDate];
}

unsigned long NSDictionary_inst_filePosixPermissions(void *id) {
	return [(NSDictionary*)id
		filePosixPermissions];
}

void* NSDictionary_inst_fileOwnerAccountID(void *id) {
	return [(NSDictionary*)id
		fileOwnerAccountID];
}

void* NSDictionary_inst_fileOwnerAccountName(void *id) {
	return [(NSDictionary*)id
		fileOwnerAccountName];
}

void* NSDictionary_inst_fileGroupOwnerAccountID(void *id) {
	return [(NSDictionary*)id
		fileGroupOwnerAccountID];
}

void* NSDictionary_inst_fileGroupOwnerAccountName(void *id) {
	return [(NSDictionary*)id
		fileGroupOwnerAccountName];
}

BOOL NSDictionary_inst_fileExtensionHidden(void *id) {
	return [(NSDictionary*)id
		fileExtensionHidden];
}

BOOL NSDictionary_inst_fileIsImmutable(void *id) {
	return [(NSDictionary*)id
		fileIsImmutable];
}

BOOL NSDictionary_inst_fileIsAppendOnly(void *id) {
	return [(NSDictionary*)id
		fileIsAppendOnly];
}

unsigned long NSDictionary_inst_fileSystemFileNumber(void *id) {
	return [(NSDictionary*)id
		fileSystemFileNumber];
}

long NSDictionary_inst_fileSystemNumber(void *id) {
	return [(NSDictionary*)id
		fileSystemNumber];
}

void* NSDictionary_inst_descriptionWithLocale_(void *id, void* locale) {
	return [(NSDictionary*)id
		descriptionWithLocale: locale];
}

void* NSDictionary_inst_descriptionWithLocale_indent_(void *id, void* locale, unsigned long level) {
	return [(NSDictionary*)id
		descriptionWithLocale: locale
		indent: level];
}

unsigned long NSDictionary_inst_count(void *id) {
	return [(NSDictionary*)id
		count];
}

void* NSDictionary_inst_allKeys(void *id) {
	return [(NSDictionary*)id
		allKeys];
}

void* NSDictionary_inst_allValues(void *id) {
	return [(NSDictionary*)id
		allValues];
}

void* NSDictionary_inst_description(void *id) {
	return [(NSDictionary*)id
		description];
}

void* NSDictionary_inst_descriptionInStringsFileFormat(void *id) {
	return [(NSDictionary*)id
		descriptionInStringsFileFormat];
}

void* NSNumber_inst_initWithBool_(void *id, BOOL value) {
	return [(NSNumber*)id
		initWithBool: value];
}

void* NSNumber_inst_initWithDouble_(void *id, double value) {
	return [(NSNumber*)id
		initWithDouble: value];
}

void* NSNumber_inst_initWithFloat_(void *id, float value) {
	return [(NSNumber*)id
		initWithFloat: value];
}

void* NSNumber_inst_initWithInt_(void *id, int value) {
	return [(NSNumber*)id
		initWithInt: value];
}

void* NSNumber_inst_initWithInteger_(void *id, long value) {
	return [(NSNumber*)id
		initWithInteger: value];
}

void* NSNumber_inst_initWithUnsignedInt_(void *id, int value) {
	return [(NSNumber*)id
		initWithUnsignedInt: value];
}

void* NSNumber_inst_initWithUnsignedInteger_(void *id, unsigned long value) {
	return [(NSNumber*)id
		initWithUnsignedInteger: value];
}

void* NSNumber_inst_descriptionWithLocale_(void *id, void* locale) {
	return [(NSNumber*)id
		descriptionWithLocale: locale];
}

BOOL NSNumber_inst_isEqualToNumber_(void *id, void* number) {
	return [(NSNumber*)id
		isEqualToNumber: number];
}

void* NSNumber_inst_init(void *id) {
	return [(NSNumber*)id
		init];
}

BOOL NSNumber_inst_boolValue(void *id) {
	return [(NSNumber*)id
		boolValue];
}

double NSNumber_inst_doubleValue(void *id) {
	return [(NSNumber*)id
		doubleValue];
}

float NSNumber_inst_floatValue(void *id) {
	return [(NSNumber*)id
		floatValue];
}

int NSNumber_inst_intValue(void *id) {
	return [(NSNumber*)id
		intValue];
}

long NSNumber_inst_integerValue(void *id) {
	return [(NSNumber*)id
		integerValue];
}

unsigned long NSNumber_inst_unsignedIntegerValue(void *id) {
	return [(NSNumber*)id
		unsignedIntegerValue];
}

int NSNumber_inst_unsignedIntValue(void *id) {
	return [(NSNumber*)id
		unsignedIntValue];
}

void* NSNumber_inst_stringValue(void *id) {
	return [(NSNumber*)id
		stringValue];
}

void NSRunLoop_inst_run(void *id) {
	[(NSRunLoop*)id
		run];
}

void NSRunLoop_inst_runUntilDate_(void *id, void* limitDate) {
	[(NSRunLoop*)id
		runUntilDate: limitDate];
}

void NSRunLoop_inst_performSelector_target_argument_order_modes_(void *id, void* aSelector, void* target, void* arg, unsigned long order, void* modes) {
	[(NSRunLoop*)id
		performSelector: aSelector
		target: target
		argument: arg
		order: order
		modes: modes];
}

void NSRunLoop_inst_cancelPerformSelector_target_argument_(void *id, void* aSelector, void* target, void* arg) {
	[(NSRunLoop*)id
		cancelPerformSelector: aSelector
		target: target
		argument: arg];
}

void NSRunLoop_inst_cancelPerformSelectorsWithTarget_(void *id, void* target) {
	[(NSRunLoop*)id
		cancelPerformSelectorsWithTarget: target];
}

void* NSRunLoop_inst_init(void *id) {
	return [(NSRunLoop*)id
		init];
}

void* NSString_inst_init(void *id) {
	return [(NSString*)id
		init];
}

void* NSString_inst_initWithBytes_length_encoding_(void *id, void* bytes, unsigned long len, unsigned long encoding) {
	return [(NSString*)id
		initWithBytes: bytes
		length: len
		encoding: encoding];
}

void* NSString_inst_initWithBytesNoCopy_length_encoding_freeWhenDone_(void *id, void* bytes, unsigned long len, unsigned long encoding, BOOL freeBuffer) {
	return [(NSString*)id
		initWithBytesNoCopy: bytes
		length: len
		encoding: encoding
		freeWhenDone: freeBuffer];
}

void* NSString_inst_initWithString_(void *id, void* aString) {
	return [(NSString*)id
		initWithString: aString];
}

void* NSString_inst_initWithData_encoding_(void *id, void* data, unsigned long encoding) {
	return [(NSString*)id
		initWithData: data
		encoding: encoding];
}

void* NSString_inst_initWithContentsOfFile_encoding_error_(void *id, void* path, unsigned long enc, void* error) {
	return [(NSString*)id
		initWithContentsOfFile: path
		encoding: enc
		error: error];
}

void* NSString_inst_initWithContentsOfURL_encoding_error_(void *id, void* url, unsigned long enc, void* error) {
	return [(NSString*)id
		initWithContentsOfURL: url
		encoding: enc
		error: error];
}

unsigned long NSString_inst_lengthOfBytesUsingEncoding_(void *id, unsigned long enc) {
	return [(NSString*)id
		lengthOfBytesUsingEncoding: enc];
}

unsigned long NSString_inst_maximumLengthOfBytesUsingEncoding_(void *id, unsigned long enc) {
	return [(NSString*)id
		maximumLengthOfBytesUsingEncoding: enc];
}

unsigned short NSString_inst_characterAtIndex_(void *id, unsigned long index) {
	return [(NSString*)id
		characterAtIndex: index];
}

BOOL NSString_inst_hasPrefix_(void *id, void* str) {
	return [(NSString*)id
		hasPrefix: str];
}

BOOL NSString_inst_hasSuffix_(void *id, void* str) {
	return [(NSString*)id
		hasSuffix: str];
}

BOOL NSString_inst_isEqualToString_(void *id, void* aString) {
	return [(NSString*)id
		isEqualToString: aString];
}

void* NSString_inst_stringByAppendingString_(void *id, void* aString) {
	return [(NSString*)id
		stringByAppendingString: aString];
}

void* NSString_inst_stringByPaddingToLength_withString_startingAtIndex_(void *id, unsigned long newLength, void* padString, unsigned long padIndex) {
	return [(NSString*)id
		stringByPaddingToLength: newLength
		withString: padString
		startingAtIndex: padIndex];
}

void* NSString_inst_componentsSeparatedByString_(void *id, void* separator) {
	return [(NSString*)id
		componentsSeparatedByString: separator];
}

void* NSString_inst_substringFromIndex_(void *id, unsigned long from) {
	return [(NSString*)id
		substringFromIndex: from];
}

void* NSString_inst_substringToIndex_(void *id, unsigned long to) {
	return [(NSString*)id
		substringToIndex: to];
}

BOOL NSString_inst_containsString_(void *id, void* str) {
	return [(NSString*)id
		containsString: str];
}

BOOL NSString_inst_localizedCaseInsensitiveContainsString_(void *id, void* str) {
	return [(NSString*)id
		localizedCaseInsensitiveContainsString: str];
}

BOOL NSString_inst_localizedStandardContainsString_(void *id, void* str) {
	return [(NSString*)id
		localizedStandardContainsString: str];
}

void* NSString_inst_stringByReplacingOccurrencesOfString_withString_(void *id, void* target, void* replacement) {
	return [(NSString*)id
		stringByReplacingOccurrencesOfString: target
		withString: replacement];
}

BOOL NSString_inst_writeToFile_atomically_encoding_error_(void *id, void* path, BOOL useAuxiliaryFile, unsigned long enc, void* error) {
	return [(NSString*)id
		writeToFile: path
		atomically: useAuxiliaryFile
		encoding: enc
		error: error];
}

BOOL NSString_inst_writeToURL_atomically_encoding_error_(void *id, void* url, BOOL useAuxiliaryFile, unsigned long enc, void* error) {
	return [(NSString*)id
		writeToURL: url
		atomically: useAuxiliaryFile
		encoding: enc
		error: error];
}

void* NSString_inst_propertyList(void *id) {
	return [(NSString*)id
		propertyList];
}

void* NSString_inst_propertyListFromStringsFileFormat(void *id) {
	return [(NSString*)id
		propertyListFromStringsFileFormat];
}

void NSString_inst_drawInRect_withAttributes_(void *id, NSRect rect, void* attrs) {
	[(NSString*)id
		drawInRect: rect
		withAttributes: attrs];
}

NSSize NSString_inst_sizeWithAttributes_(void *id, void* attrs) {
	return [(NSString*)id
		sizeWithAttributes: attrs];
}

void* NSString_inst_variantFittingPresentationWidth_(void *id, long width) {
	return [(NSString*)id
		variantFittingPresentationWidth: width];
}

BOOL NSString_inst_canBeConvertedToEncoding_(void *id, unsigned long encoding) {
	return [(NSString*)id
		canBeConvertedToEncoding: encoding];
}

void* NSString_inst_dataUsingEncoding_(void *id, unsigned long encoding) {
	return [(NSString*)id
		dataUsingEncoding: encoding];
}

void* NSString_inst_dataUsingEncoding_allowLossyConversion_(void *id, unsigned long encoding, BOOL lossy) {
	return [(NSString*)id
		dataUsingEncoding: encoding
		allowLossyConversion: lossy];
}

unsigned long NSString_inst_completePathIntoString_caseSensitive_matchesIntoArray_filterTypes_(void *id, void* outputName, BOOL flag, void* outputArray, void* filterTypes) {
	return [(NSString*)id
		completePathIntoString: outputName
		caseSensitive: flag
		matchesIntoArray: outputArray
		filterTypes: filterTypes];
}

void* NSString_inst_stringByAppendingPathComponent_(void *id, void* str) {
	return [(NSString*)id
		stringByAppendingPathComponent: str];
}

void* NSString_inst_stringByAppendingPathExtension_(void *id, void* str) {
	return [(NSString*)id
		stringByAppendingPathExtension: str];
}

void* NSString_inst_stringsByAppendingPaths_(void *id, void* paths) {
	return [(NSString*)id
		stringsByAppendingPaths: paths];
}

unsigned long NSString_inst_length(void *id) {
	return [(NSString*)id
		length];
}

unsigned long NSString_inst_hash(void *id) {
	return [(NSString*)id
		hash];
}

void* NSString_inst_lowercaseString(void *id) {
	return [(NSString*)id
		lowercaseString];
}

void* NSString_inst_localizedLowercaseString(void *id) {
	return [(NSString*)id
		localizedLowercaseString];
}

void* NSString_inst_uppercaseString(void *id) {
	return [(NSString*)id
		uppercaseString];
}

void* NSString_inst_localizedUppercaseString(void *id) {
	return [(NSString*)id
		localizedUppercaseString];
}

void* NSString_inst_capitalizedString(void *id) {
	return [(NSString*)id
		capitalizedString];
}

void* NSString_inst_localizedCapitalizedString(void *id) {
	return [(NSString*)id
		localizedCapitalizedString];
}

void* NSString_inst_decomposedStringWithCanonicalMapping(void *id) {
	return [(NSString*)id
		decomposedStringWithCanonicalMapping];
}

void* NSString_inst_decomposedStringWithCompatibilityMapping(void *id) {
	return [(NSString*)id
		decomposedStringWithCompatibilityMapping];
}

void* NSString_inst_precomposedStringWithCanonicalMapping(void *id) {
	return [(NSString*)id
		precomposedStringWithCanonicalMapping];
}

void* NSString_inst_precomposedStringWithCompatibilityMapping(void *id) {
	return [(NSString*)id
		precomposedStringWithCompatibilityMapping];
}

double NSString_inst_doubleValue(void *id) {
	return [(NSString*)id
		doubleValue];
}

float NSString_inst_floatValue(void *id) {
	return [(NSString*)id
		floatValue];
}

int NSString_inst_intValue(void *id) {
	return [(NSString*)id
		intValue];
}

long NSString_inst_integerValue(void *id) {
	return [(NSString*)id
		integerValue];
}

BOOL NSString_inst_boolValue(void *id) {
	return [(NSString*)id
		boolValue];
}

void* NSString_inst_description(void *id) {
	return [(NSString*)id
		description];
}

unsigned long NSString_inst_fastestEncoding(void *id) {
	return [(NSString*)id
		fastestEncoding];
}

unsigned long NSString_inst_smallestEncoding(void *id) {
	return [(NSString*)id
		smallestEncoding];
}

void* NSString_inst_pathComponents(void *id) {
	return [(NSString*)id
		pathComponents];
}

BOOL NSString_inst_isAbsolutePath(void *id) {
	return [(NSString*)id
		isAbsolutePath];
}

void* NSString_inst_lastPathComponent(void *id) {
	return [(NSString*)id
		lastPathComponent];
}

void* NSString_inst_pathExtension(void *id) {
	return [(NSString*)id
		pathExtension];
}

void* NSString_inst_stringByAbbreviatingWithTildeInPath(void *id) {
	return [(NSString*)id
		stringByAbbreviatingWithTildeInPath];
}

void* NSString_inst_stringByDeletingLastPathComponent(void *id) {
	return [(NSString*)id
		stringByDeletingLastPathComponent];
}

void* NSString_inst_stringByDeletingPathExtension(void *id) {
	return [(NSString*)id
		stringByDeletingPathExtension];
}

void* NSString_inst_stringByExpandingTildeInPath(void *id) {
	return [(NSString*)id
		stringByExpandingTildeInPath];
}

void* NSString_inst_stringByResolvingSymlinksInPath(void *id) {
	return [(NSString*)id
		stringByResolvingSymlinksInPath];
}

void* NSString_inst_stringByStandardizingPath(void *id) {
	return [(NSString*)id
		stringByStandardizingPath];
}

void* NSString_inst_stringByRemovingPercentEncoding(void *id) {
	return [(NSString*)id
		stringByRemovingPercentEncoding];
}

void* NSError_inst_initWithDomain_code_userInfo_(void *id, void* domain, long code, void* dict) {
	return [(NSError*)id
		initWithDomain: domain
		code: code
		userInfo: dict];
}

void* NSError_inst_init(void *id) {
	return [(NSError*)id
		init];
}

long NSError_inst_code(void *id) {
	return [(NSError*)id
		code];
}

void* NSError_inst_domain(void *id) {
	return [(NSError*)id
		domain];
}

void* NSError_inst_userInfo(void *id) {
	return [(NSError*)id
		userInfo];
}

void* NSError_inst_localizedDescription(void *id) {
	return [(NSError*)id
		localizedDescription];
}

void* NSError_inst_localizedRecoveryOptions(void *id) {
	return [(NSError*)id
		localizedRecoveryOptions];
}

void* NSError_inst_localizedRecoverySuggestion(void *id) {
	return [(NSError*)id
		localizedRecoverySuggestion];
}

void* NSError_inst_localizedFailureReason(void *id) {
	return [(NSError*)id
		localizedFailureReason];
}

void* NSError_inst_recoveryAttempter(void *id) {
	return [(NSError*)id
		recoveryAttempter];
}

void* NSError_inst_helpAnchor(void *id) {
	return [(NSError*)id
		helpAnchor];
}

void* NSError_inst_underlyingErrors(void *id) {
	return [(NSError*)id
		underlyingErrors];
}

void* NSThread_inst_init(void *id) {
	return [(NSThread*)id
		init];
}

void* NSThread_inst_initWithTarget_selector_object_(void *id, void* target, void* selector, void* argument) {
	return [(NSThread*)id
		initWithTarget: target
		selector: selector
		object: argument];
}

void NSThread_inst_start(void *id) {
	[(NSThread*)id
		start];
}

void NSThread_inst_main(void *id) {
	[(NSThread*)id
		main];
}

void NSThread_inst_cancel(void *id) {
	[(NSThread*)id
		cancel];
}

BOOL NSThread_inst_isExecuting(void *id) {
	return [(NSThread*)id
		isExecuting];
}

BOOL NSThread_inst_isFinished(void *id) {
	return [(NSThread*)id
		isFinished];
}

BOOL NSThread_inst_isCancelled(void *id) {
	return [(NSThread*)id
		isCancelled];
}

BOOL NSThread_inst_isMainThread(void *id) {
	return [(NSThread*)id
		isMainThread];
}

void* NSThread_inst_name(void *id) {
	return [(NSThread*)id
		name];
}

void NSThread_inst_setName_(void *id, void* value) {
	[(NSThread*)id
		setName: value];
}

unsigned long NSThread_inst_stackSize(void *id) {
	return [(NSThread*)id
		stackSize];
}

void NSThread_inst_setStackSize_(void *id, unsigned long value) {
	[(NSThread*)id
		setStackSize: value];
}

double NSThread_inst_threadPriority(void *id) {
	return [(NSThread*)id
		threadPriority];
}

void NSThread_inst_setThreadPriority_(void *id, double value) {
	[(NSThread*)id
		setThreadPriority: value];
}

void* NSURL_inst_initWithString_(void *id, void* URLString) {
	return [(NSURL*)id
		initWithString: URLString];
}

void* NSURL_inst_initWithString_relativeToURL_(void *id, void* URLString, void* baseURL) {
	return [(NSURL*)id
		initWithString: URLString
		relativeToURL: baseURL];
}

void* NSURL_inst_initFileURLWithPath_isDirectory_(void *id, void* path, BOOL isDir) {
	return [(NSURL*)id
		initFileURLWithPath: path
		isDirectory: isDir];
}

void* NSURL_inst_initFileURLWithPath_relativeToURL_(void *id, void* path, void* baseURL) {
	return [(NSURL*)id
		initFileURLWithPath: path
		relativeToURL: baseURL];
}

void* NSURL_inst_initFileURLWithPath_isDirectory_relativeToURL_(void *id, void* path, BOOL isDir, void* baseURL) {
	return [(NSURL*)id
		initFileURLWithPath: path
		isDirectory: isDir
		relativeToURL: baseURL];
}

void* NSURL_inst_initFileURLWithPath_(void *id, void* path) {
	return [(NSURL*)id
		initFileURLWithPath: path];
}

void* NSURL_inst_initAbsoluteURLWithDataRepresentation_relativeToURL_(void *id, void* data, void* baseURL) {
	return [(NSURL*)id
		initAbsoluteURLWithDataRepresentation: data
		relativeToURL: baseURL];
}

void* NSURL_inst_initWithDataRepresentation_relativeToURL_(void *id, void* data, void* baseURL) {
	return [(NSURL*)id
		initWithDataRepresentation: data
		relativeToURL: baseURL];
}

BOOL NSURL_inst_isEqual_(void *id, void* anObject) {
	return [(NSURL*)id
		isEqual: anObject];
}

BOOL NSURL_inst_checkResourceIsReachableAndReturnError_(void *id, void* error) {
	return [(NSURL*)id
		checkResourceIsReachableAndReturnError: error];
}

BOOL NSURL_inst_isFileReferenceURL(void *id) {
	return [(NSURL*)id
		isFileReferenceURL];
}

void* NSURL_inst_resourceValuesForKeys_error_(void *id, void* keys, void* error) {
	return [(NSURL*)id
		resourceValuesForKeys: keys
		error: error];
}

BOOL NSURL_inst_setResourceValues_error_(void *id, void* keyedValues, void* error) {
	return [(NSURL*)id
		setResourceValues: keyedValues
		error: error];
}

void NSURL_inst_removeAllCachedResourceValues(void *id) {
	[(NSURL*)id
		removeAllCachedResourceValues];
}

void* NSURL_inst_fileReferenceURL(void *id) {
	return [(NSURL*)id
		fileReferenceURL];
}

void* NSURL_inst_URLByAppendingPathComponent_(void *id, void* pathComponent) {
	return [(NSURL*)id
		URLByAppendingPathComponent: pathComponent];
}

void* NSURL_inst_URLByAppendingPathComponent_isDirectory_(void *id, void* pathComponent, BOOL isDirectory) {
	return [(NSURL*)id
		URLByAppendingPathComponent: pathComponent
		isDirectory: isDirectory];
}

void* NSURL_inst_URLByAppendingPathExtension_(void *id, void* pathExtension) {
	return [(NSURL*)id
		URLByAppendingPathExtension: pathExtension];
}

BOOL NSURL_inst_startAccessingSecurityScopedResource(void *id) {
	return [(NSURL*)id
		startAccessingSecurityScopedResource];
}

void NSURL_inst_stopAccessingSecurityScopedResource(void *id) {
	[(NSURL*)id
		stopAccessingSecurityScopedResource];
}

BOOL NSURL_inst_checkPromisedItemIsReachableAndReturnError_(void *id, void* error) {
	return [(NSURL*)id
		checkPromisedItemIsReachableAndReturnError: error];
}

void* NSURL_inst_promisedItemResourceValuesForKeys_error_(void *id, void* keys, void* error) {
	return [(NSURL*)id
		promisedItemResourceValuesForKeys: keys
		error: error];
}

void* NSURL_inst_init(void *id) {
	return [(NSURL*)id
		init];
}

void* NSURL_inst_dataRepresentation(void *id) {
	return [(NSURL*)id
		dataRepresentation];
}

BOOL NSURL_inst_isFileURL(void *id) {
	return [(NSURL*)id
		isFileURL];
}

void* NSURL_inst_absoluteString(void *id) {
	return [(NSURL*)id
		absoluteString];
}

void* NSURL_inst_absoluteURL(void *id) {
	return [(NSURL*)id
		absoluteURL];
}

void* NSURL_inst_baseURL(void *id) {
	return [(NSURL*)id
		baseURL];
}

void* NSURL_inst_fragment(void *id) {
	return [(NSURL*)id
		fragment];
}

void* NSURL_inst_host(void *id) {
	return [(NSURL*)id
		host];
}

void* NSURL_inst_lastPathComponent(void *id) {
	return [(NSURL*)id
		lastPathComponent];
}

void* NSURL_inst_password(void *id) {
	return [(NSURL*)id
		password];
}

void* NSURL_inst_path(void *id) {
	return [(NSURL*)id
		path];
}

void* NSURL_inst_pathComponents(void *id) {
	return [(NSURL*)id
		pathComponents];
}

void* NSURL_inst_pathExtension(void *id) {
	return [(NSURL*)id
		pathExtension];
}

void* NSURL_inst_port(void *id) {
	return [(NSURL*)id
		port];
}

void* NSURL_inst_query(void *id) {
	return [(NSURL*)id
		query];
}

void* NSURL_inst_relativePath(void *id) {
	return [(NSURL*)id
		relativePath];
}

void* NSURL_inst_relativeString(void *id) {
	return [(NSURL*)id
		relativeString];
}

void* NSURL_inst_resourceSpecifier(void *id) {
	return [(NSURL*)id
		resourceSpecifier];
}

void* NSURL_inst_scheme(void *id) {
	return [(NSURL*)id
		scheme];
}

void* NSURL_inst_standardizedURL(void *id) {
	return [(NSURL*)id
		standardizedURL];
}

void* NSURL_inst_user(void *id) {
	return [(NSURL*)id
		user];
}

void* NSURL_inst_filePathURL(void *id) {
	return [(NSURL*)id
		filePathURL];
}

void* NSURL_inst_URLByDeletingLastPathComponent(void *id) {
	return [(NSURL*)id
		URLByDeletingLastPathComponent];
}

void* NSURL_inst_URLByDeletingPathExtension(void *id) {
	return [(NSURL*)id
		URLByDeletingPathExtension];
}

void* NSURL_inst_URLByResolvingSymlinksInPath(void *id) {
	return [(NSURL*)id
		URLByResolvingSymlinksInPath];
}

void* NSURL_inst_URLByStandardizingPath(void *id) {
	return [(NSURL*)id
		URLByStandardizingPath];
}

BOOL NSURL_inst_hasDirectoryPath(void *id) {
	return [(NSURL*)id
		hasDirectoryPath];
}

void* NSURLRequest_inst_initWithURL_(void *id, void* URL) {
	return [(NSURLRequest*)id
		initWithURL: URL];
}

void* NSURLRequest_inst_valueForHTTPHeaderField_(void *id, void* field) {
	return [(NSURLRequest*)id
		valueForHTTPHeaderField: field];
}

void* NSURLRequest_inst_init(void *id) {
	return [(NSURLRequest*)id
		init];
}

void* NSURLRequest_inst_HTTPMethod(void *id) {
	return [(NSURLRequest*)id
		HTTPMethod];
}

void* NSURLRequest_inst_URL(void *id) {
	return [(NSURLRequest*)id
		URL];
}

void* NSURLRequest_inst_HTTPBody(void *id) {
	return [(NSURLRequest*)id
		HTTPBody];
}

void* NSURLRequest_inst_mainDocumentURL(void *id) {
	return [(NSURLRequest*)id
		mainDocumentURL];
}

void* NSURLRequest_inst_allHTTPHeaderFields(void *id) {
	return [(NSURLRequest*)id
		allHTTPHeaderFields];
}

NSTimeInterval NSURLRequest_inst_timeoutInterval(void *id) {
	return [(NSURLRequest*)id
		timeoutInterval];
}

BOOL NSURLRequest_inst_HTTPShouldHandleCookies(void *id) {
	return [(NSURLRequest*)id
		HTTPShouldHandleCookies];
}

BOOL NSURLRequest_inst_HTTPShouldUsePipelining(void *id) {
	return [(NSURLRequest*)id
		HTTPShouldUsePipelining];
}

BOOL NSURLRequest_inst_allowsCellularAccess(void *id) {
	return [(NSURLRequest*)id
		allowsCellularAccess];
}

BOOL NSURLRequest_inst_allowsConstrainedNetworkAccess(void *id) {
	return [(NSURLRequest*)id
		allowsConstrainedNetworkAccess];
}

BOOL NSURLRequest_inst_allowsExpensiveNetworkAccess(void *id) {
	return [(NSURLRequest*)id
		allowsExpensiveNetworkAccess];
}

BOOL NSURLRequest_inst_assumesHTTP3Capable(void *id) {
	return [(NSURLRequest*)id
		assumesHTTP3Capable];
}

void* NSNotification_inst_init(void *id) {
	return [(NSNotification*)id
		init];
}

void* NSNotification_inst_object_(void *id) {
	return [(NSNotification*)id
		object_];
}

void* NSNotification_inst_userInfo(void *id) {
	return [(NSNotification*)id
		userInfo];
}

void NSOperationQueue_inst_addOperations_waitUntilFinished_(void *id, void* ops, BOOL wait) {
	[(NSOperationQueue*)id
		addOperations: ops
		waitUntilFinished: wait];
}

void NSOperationQueue_inst_cancelAllOperations(void *id) {
	[(NSOperationQueue*)id
		cancelAllOperations];
}

void NSOperationQueue_inst_waitUntilAllOperationsAreFinished(void *id) {
	[(NSOperationQueue*)id
		waitUntilAllOperationsAreFinished];
}

void* NSOperationQueue_inst_init(void *id) {
	return [(NSOperationQueue*)id
		init];
}

long NSOperationQueue_inst_maxConcurrentOperationCount(void *id) {
	return [(NSOperationQueue*)id
		maxConcurrentOperationCount];
}

void NSOperationQueue_inst_setMaxConcurrentOperationCount_(void *id, long value) {
	[(NSOperationQueue*)id
		setMaxConcurrentOperationCount: value];
}

BOOL NSOperationQueue_inst_isSuspended(void *id) {
	return [(NSOperationQueue*)id
		isSuspended];
}

void NSOperationQueue_inst_setSuspended_(void *id, BOOL value) {
	[(NSOperationQueue*)id
		setSuspended: value];
}

void* NSOperationQueue_inst_name(void *id) {
	return [(NSOperationQueue*)id
		name];
}

void NSOperationQueue_inst_setName_(void *id, void* value) {
	[(NSOperationQueue*)id
		setName: value];
}

void NSNotificationCenter_inst_addObserver_selector_name_object_(void *id, void* observer, void* aSelector, void* aName, void* anObject) {
	[(NSNotificationCenter*)id
		addObserver: observer
		selector: aSelector
		name: aName
		object: anObject];
}

void NSNotificationCenter_inst_removeObserver_name_object_(void *id, void* observer, void* aName, void* anObject) {
	[(NSNotificationCenter*)id
		removeObserver: observer
		name: aName
		object: anObject];
}

void NSNotificationCenter_inst_removeObserver_(void *id, void* observer) {
	[(NSNotificationCenter*)id
		removeObserver: observer];
}

void NSNotificationCenter_inst_postNotification_(void *id, void* notification) {
	[(NSNotificationCenter*)id
		postNotification: notification];
}

void NSNotificationCenter_inst_postNotificationName_object_userInfo_(void *id, void* aName, void* anObject, void* aUserInfo) {
	[(NSNotificationCenter*)id
		postNotificationName: aName
		object: anObject
		userInfo: aUserInfo];
}

void NSNotificationCenter_inst_postNotificationName_object_(void *id, void* aName, void* anObject) {
	[(NSNotificationCenter*)id
		postNotificationName: aName
		object: anObject];
}

void* NSNotificationCenter_inst_init(void *id) {
	return [(NSNotificationCenter*)id
		init];
}


BOOL core_objc_bool_true = YES;
BOOL core_objc_bool_false = NO;

*/
import "C"

func convertObjCBoolToGo(b C.BOOL) bool {
	// NOTE: the prefix here is used to namespace these since the linker will
	// otherwise report a "duplicate symbol" because the C functions have the
	// same name.
	return bool(C.core_convertObjCBool(b))
}

func convertToObjCBool(b bool) C.BOOL {
	if b {
		return C.core_objc_bool_true
	}
	return C.core_objc_bool_false
}

func CALayer_alloc() (
	r0 CALayer,
) {
	ret := C.CALayer_type_alloc()
	r0 = CALayer_fromPointer(ret)
	return
}

func CALayer_layer() (
	r0 CALayer,
) {
	ret := C.CALayer_type_layer()
	r0 = CALayer_fromPointer(ret)
	return
}

func CALayer_layerWithRemoteClientId_(
	client_id uint32,
) (
	r0 CALayer,
) {
	ret := C.CALayer_type_layerWithRemoteClientId_(
		C.uint32_t(client_id),
	)
	r0 = CALayer_fromPointer(ret)
	return
}

func CALayer_needsDisplayForKey_(
	key NSStringRef,
) (
	r0 bool,
) {
	ret := C.CALayer_type_needsDisplayForKey_(
		objc.RefPointer(key),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func CALayer_defaultActionForKey_(
	event NSStringRef,
) (
	r0 objc.Object,
) {
	ret := C.CALayer_type_defaultActionForKey_(
		objc.RefPointer(event),
	)
	r0 = objc.Object_fromPointer(ret)
	return
}

func CALayer_defaultValueForKey_(
	key NSStringRef,
) (
	r0 objc.Object,
) {
	ret := C.CALayer_type_defaultValueForKey_(
		objc.RefPointer(key),
	)
	r0 = objc.Object_fromPointer(ret)
	return
}

func NSArray_alloc() (
	r0 NSArray,
) {
	ret := C.NSArray_type_alloc()
	r0 = NSArray_fromPointer(ret)
	return
}

func NSArray_array() (
	r0 NSArray,
) {
	ret := C.NSArray_type_array()
	r0 = NSArray_fromPointer(ret)
	return
}

func NSArray_arrayWithArray_(
	array NSArrayRef,
) (
	r0 NSArray,
) {
	ret := C.NSArray_type_arrayWithArray_(
		objc.RefPointer(array),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func NSArray_arrayWithObject_(
	anObject objc.Ref,
) (
	r0 NSArray,
) {
	ret := C.NSArray_type_arrayWithObject_(
		objc.RefPointer(anObject),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func NSArray_arrayWithContentsOfURL_error_(
	url NSURLRef,
	error NSErrorRef,
) (
	r0 NSArray,
) {
	ret := C.NSArray_type_arrayWithContentsOfURL_error_(
		objc.RefPointer(url),
		objc.RefPointer(error),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func NSAttributedString_alloc() (
	r0 NSAttributedString,
) {
	ret := C.NSAttributedString_type_alloc()
	r0 = NSAttributedString_fromPointer(ret)
	return
}

func NSAttributedString_textTypes() (
	r0 NSArray,
) {
	ret := C.NSAttributedString_type_textTypes()
	r0 = NSArray_fromPointer(ret)
	return
}

func NSAttributedString_textUnfilteredTypes() (
	r0 NSArray,
) {
	ret := C.NSAttributedString_type_textUnfilteredTypes()
	r0 = NSArray_fromPointer(ret)
	return
}

func NSData_alloc() (
	r0 NSData,
) {
	ret := C.NSData_type_alloc()
	r0 = NSData_fromPointer(ret)
	return
}

func NSData_data() (
	r0 NSData,
) {
	ret := C.NSData_type_data()
	r0 = NSData_fromPointer(ret)
	return
}

func NSData_dataWithBytes_length_(
	bytes unsafe.Pointer,
	length NSUInteger,
) (
	r0 NSData,
) {
	ret := C.NSData_type_dataWithBytes_length_(
		bytes,
		C.ulong(length),
	)
	r0 = NSData_fromPointer(ret)
	return
}

func NSData_dataWithBytesNoCopy_length_(
	bytes unsafe.Pointer,
	length NSUInteger,
) (
	r0 NSData,
) {
	ret := C.NSData_type_dataWithBytesNoCopy_length_(
		bytes,
		C.ulong(length),
	)
	r0 = NSData_fromPointer(ret)
	return
}

func NSData_dataWithBytesNoCopy_length_freeWhenDone_(
	bytes unsafe.Pointer,
	length NSUInteger,
	b bool,
) (
	r0 NSData,
) {
	ret := C.NSData_type_dataWithBytesNoCopy_length_freeWhenDone_(
		bytes,
		C.ulong(length),
		convertToObjCBool(b),
	)
	r0 = NSData_fromPointer(ret)
	return
}

func NSData_dataWithData_(
	data NSDataRef,
) (
	r0 NSData,
) {
	ret := C.NSData_type_dataWithData_(
		objc.RefPointer(data),
	)
	r0 = NSData_fromPointer(ret)
	return
}

func NSData_dataWithContentsOfFile_(
	path NSStringRef,
) (
	r0 NSData,
) {
	ret := C.NSData_type_dataWithContentsOfFile_(
		objc.RefPointer(path),
	)
	r0 = NSData_fromPointer(ret)
	return
}

func NSData_dataWithContentsOfURL_(
	url NSURLRef,
) (
	r0 NSData,
) {
	ret := C.NSData_type_dataWithContentsOfURL_(
		objc.RefPointer(url),
	)
	r0 = NSData_fromPointer(ret)
	return
}

func NSDate_alloc() (
	r0 NSDate,
) {
	ret := C.NSDate_type_alloc()
	r0 = NSDate_fromPointer(ret)
	return
}

func NSDate_date() (
	r0 NSDate,
) {
	ret := C.NSDate_type_date()
	r0 = NSDate_fromPointer(ret)
	return
}

func NSDate_distantFuture() (
	r0 NSDate,
) {
	ret := C.NSDate_type_distantFuture()
	r0 = NSDate_fromPointer(ret)
	return
}

func NSDate_distantPast() (
	r0 NSDate,
) {
	ret := C.NSDate_type_distantPast()
	r0 = NSDate_fromPointer(ret)
	return
}

func NSDate_now() (
	r0 NSDate,
) {
	ret := C.NSDate_type_now()
	r0 = NSDate_fromPointer(ret)
	return
}

func NSDictionary_alloc() (
	r0 NSDictionary,
) {
	ret := C.NSDictionary_type_alloc()
	r0 = NSDictionary_fromPointer(ret)
	return
}

func NSDictionary_dictionary() (
	r0 NSDictionary,
) {
	ret := C.NSDictionary_type_dictionary()
	r0 = NSDictionary_fromPointer(ret)
	return
}

func NSDictionary_dictionaryWithObjects_forKeys_(
	objects NSArrayRef,
	keys NSArrayRef,
) (
	r0 NSDictionary,
) {
	ret := C.NSDictionary_type_dictionaryWithObjects_forKeys_(
		objc.RefPointer(objects),
		objc.RefPointer(keys),
	)
	r0 = NSDictionary_fromPointer(ret)
	return
}

func NSDictionary_dictionaryWithObject_forKey_(
	object objc.Ref,
	key objc.Ref,
) (
	r0 NSDictionary,
) {
	ret := C.NSDictionary_type_dictionaryWithObject_forKey_(
		objc.RefPointer(object),
		objc.RefPointer(key),
	)
	r0 = NSDictionary_fromPointer(ret)
	return
}

func NSDictionary_dictionaryWithDictionary_(
	dict NSDictionaryRef,
) (
	r0 NSDictionary,
) {
	ret := C.NSDictionary_type_dictionaryWithDictionary_(
		objc.RefPointer(dict),
	)
	r0 = NSDictionary_fromPointer(ret)
	return
}

func NSDictionary_dictionaryWithContentsOfURL_error_(
	url NSURLRef,
	error NSErrorRef,
) (
	r0 NSDictionary,
) {
	ret := C.NSDictionary_type_dictionaryWithContentsOfURL_error_(
		objc.RefPointer(url),
		objc.RefPointer(error),
	)
	r0 = NSDictionary_fromPointer(ret)
	return
}

func NSDictionary_sharedKeySetForKeys_(
	keys NSArrayRef,
) (
	r0 objc.Object,
) {
	ret := C.NSDictionary_type_sharedKeySetForKeys_(
		objc.RefPointer(keys),
	)
	r0 = objc.Object_fromPointer(ret)
	return
}

func NSNumber_alloc() (
	r0 NSNumber,
) {
	ret := C.NSNumber_type_alloc()
	r0 = NSNumber_fromPointer(ret)
	return
}

func NSNumber_numberWithBool_(
	value bool,
) (
	r0 NSNumber,
) {
	ret := C.NSNumber_type_numberWithBool_(
		convertToObjCBool(value),
	)
	r0 = NSNumber_fromPointer(ret)
	return
}

func NSNumber_numberWithDouble_(
	value float64,
) (
	r0 NSNumber,
) {
	ret := C.NSNumber_type_numberWithDouble_(
		C.double(value),
	)
	r0 = NSNumber_fromPointer(ret)
	return
}

func NSNumber_numberWithFloat_(
	value float32,
) (
	r0 NSNumber,
) {
	ret := C.NSNumber_type_numberWithFloat_(
		C.float(value),
	)
	r0 = NSNumber_fromPointer(ret)
	return
}

func NSNumber_numberWithInt_(
	value int32,
) (
	r0 NSNumber,
) {
	ret := C.NSNumber_type_numberWithInt_(
		C.int(value),
	)
	r0 = NSNumber_fromPointer(ret)
	return
}

func NSNumber_numberWithInteger_(
	value NSInteger,
) (
	r0 NSNumber,
) {
	ret := C.NSNumber_type_numberWithInteger_(
		C.long(value),
	)
	r0 = NSNumber_fromPointer(ret)
	return
}

func NSNumber_numberWithUnsignedInt_(
	value int32,
) (
	r0 NSNumber,
) {
	ret := C.NSNumber_type_numberWithUnsignedInt_(
		C.int(value),
	)
	r0 = NSNumber_fromPointer(ret)
	return
}

func NSNumber_numberWithUnsignedInteger_(
	value NSUInteger,
) (
	r0 NSNumber,
) {
	ret := C.NSNumber_type_numberWithUnsignedInteger_(
		C.ulong(value),
	)
	r0 = NSNumber_fromPointer(ret)
	return
}

func NSRunLoop_alloc() (
	r0 NSRunLoop,
) {
	ret := C.NSRunLoop_type_alloc()
	r0 = NSRunLoop_fromPointer(ret)
	return
}

func NSRunLoop_currentRunLoop() (
	r0 NSRunLoop,
) {
	ret := C.NSRunLoop_type_currentRunLoop()
	r0 = NSRunLoop_fromPointer(ret)
	return
}

func NSRunLoop_mainRunLoop() (
	r0 NSRunLoop,
) {
	ret := C.NSRunLoop_type_mainRunLoop()
	r0 = NSRunLoop_fromPointer(ret)
	return
}

func NSString_alloc() (
	r0 NSString,
) {
	ret := C.NSString_type_alloc()
	r0 = NSString_fromPointer(ret)
	return
}

func NSString_string() (
	r0 NSString,
) {
	ret := C.NSString_type_string()
	r0 = NSString_fromPointer(ret)
	return
}

func NSString_localizedUserNotificationStringForKey_arguments_(
	key NSStringRef,
	arguments NSArrayRef,
) (
	r0 NSString,
) {
	ret := C.NSString_type_localizedUserNotificationStringForKey_arguments_(
		objc.RefPointer(key),
		objc.RefPointer(arguments),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func NSString_stringWithString_(
	string NSStringRef,
) (
	r0 NSString,
) {
	ret := C.NSString_type_stringWithString_(
		objc.RefPointer(string),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func NSString_stringWithContentsOfFile_encoding_error_(
	path NSStringRef,
	enc NSStringEncoding,
	error NSErrorRef,
) (
	r0 NSString,
) {
	ret := C.NSString_type_stringWithContentsOfFile_encoding_error_(
		objc.RefPointer(path),
		C.ulong(enc),
		objc.RefPointer(error),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func NSString_stringWithContentsOfURL_encoding_error_(
	url NSURLRef,
	enc NSStringEncoding,
	error NSErrorRef,
) (
	r0 NSString,
) {
	ret := C.NSString_type_stringWithContentsOfURL_encoding_error_(
		objc.RefPointer(url),
		C.ulong(enc),
		objc.RefPointer(error),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func NSString_localizedNameOfStringEncoding_(
	encoding NSStringEncoding,
) (
	r0 NSString,
) {
	ret := C.NSString_type_localizedNameOfStringEncoding_(
		C.ulong(encoding),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func NSString_pathWithComponents_(
	components NSArrayRef,
) (
	r0 NSString,
) {
	ret := C.NSString_type_pathWithComponents_(
		objc.RefPointer(components),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func NSString_defaultCStringEncoding() (
	r0 NSStringEncoding,
) {
	ret := C.NSString_type_defaultCStringEncoding()
	r0 = NSStringEncoding(ret)
	return
}

func NSError_alloc() (
	r0 NSError,
) {
	ret := C.NSError_type_alloc()
	r0 = NSError_fromPointer(ret)
	return
}

func NSError_errorWithDomain_code_userInfo_(
	domain NSStringRef,
	code NSInteger,
	dict NSDictionaryRef,
) (
	r0 NSError,
) {
	ret := C.NSError_type_errorWithDomain_code_userInfo_(
		objc.RefPointer(domain),
		C.long(code),
		objc.RefPointer(dict),
	)
	r0 = NSError_fromPointer(ret)
	return
}

func NSThread_alloc() (
	r0 NSThread,
) {
	ret := C.NSThread_type_alloc()
	r0 = NSThread_fromPointer(ret)
	return
}

func NSThread_detachNewThreadSelector_toTarget_withObject_(
	selector objc.Selector,
	target objc.Ref,
	argument objc.Ref,
) {
	C.NSThread_type_detachNewThreadSelector_toTarget_withObject_(
		selector.SelectorAddress(),
		objc.RefPointer(target),
		objc.RefPointer(argument),
	)
	return
}

func NSThread_sleepUntilDate_(
	date NSDateRef,
) {
	C.NSThread_type_sleepUntilDate_(
		objc.RefPointer(date),
	)
	return
}

func NSThread_sleepForTimeInterval_(
	ti float64,
) {
	C.NSThread_type_sleepForTimeInterval_(
		C.NSTimeInterval(ti),
	)
	return
}

func NSThread_exit() {
	C.NSThread_type_exit()
	return
}

func NSThread_isMultiThreaded() (
	r0 bool,
) {
	ret := C.NSThread_type_isMultiThreaded()
	r0 = convertObjCBoolToGo(ret)
	return
}

func NSThread_threadPriority() (
	r0 float64,
) {
	ret := C.NSThread_type_threadPriority()
	r0 = float64(ret)
	return
}

func NSThread_setThreadPriority_(
	p float64,
) (
	r0 bool,
) {
	ret := C.NSThread_type_setThreadPriority_(
		C.double(p),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func NSThread_isMainThread() (
	r0 bool,
) {
	ret := C.NSThread_type_isMainThread()
	r0 = convertObjCBoolToGo(ret)
	return
}

func NSThread_mainThread() (
	r0 NSThread,
) {
	ret := C.NSThread_type_mainThread()
	r0 = NSThread_fromPointer(ret)
	return
}

func NSThread_currentThread() (
	r0 NSThread,
) {
	ret := C.NSThread_type_currentThread()
	r0 = NSThread_fromPointer(ret)
	return
}

func NSThread_callStackReturnAddresses() (
	r0 NSArray,
) {
	ret := C.NSThread_type_callStackReturnAddresses()
	r0 = NSArray_fromPointer(ret)
	return
}

func NSThread_callStackSymbols() (
	r0 NSArray,
) {
	ret := C.NSThread_type_callStackSymbols()
	r0 = NSArray_fromPointer(ret)
	return
}

func NSURL_alloc() (
	r0 NSURL,
) {
	ret := C.NSURL_type_alloc()
	r0 = NSURL_fromPointer(ret)
	return
}

func NSURL_URLWithString_(
	URLString NSStringRef,
) (
	r0 NSURL,
) {
	ret := C.NSURL_type_URLWithString_(
		objc.RefPointer(URLString),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func NSURL_URLWithString_relativeToURL_(
	URLString NSStringRef,
	baseURL NSURLRef,
) (
	r0 NSURL,
) {
	ret := C.NSURL_type_URLWithString_relativeToURL_(
		objc.RefPointer(URLString),
		objc.RefPointer(baseURL),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func NSURL_fileURLWithPath_isDirectory_(
	path NSStringRef,
	isDir bool,
) (
	r0 NSURL,
) {
	ret := C.NSURL_type_fileURLWithPath_isDirectory_(
		objc.RefPointer(path),
		convertToObjCBool(isDir),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func NSURL_fileURLWithPath_relativeToURL_(
	path NSStringRef,
	baseURL NSURLRef,
) (
	r0 NSURL,
) {
	ret := C.NSURL_type_fileURLWithPath_relativeToURL_(
		objc.RefPointer(path),
		objc.RefPointer(baseURL),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func NSURL_fileURLWithPath_isDirectory_relativeToURL_(
	path NSStringRef,
	isDir bool,
	baseURL NSURLRef,
) (
	r0 NSURL,
) {
	ret := C.NSURL_type_fileURLWithPath_isDirectory_relativeToURL_(
		objc.RefPointer(path),
		convertToObjCBool(isDir),
		objc.RefPointer(baseURL),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func NSURL_fileURLWithPath_(
	path NSStringRef,
) (
	r0 NSURL,
) {
	ret := C.NSURL_type_fileURLWithPath_(
		objc.RefPointer(path),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func NSURL_fileURLWithPathComponents_(
	components NSArrayRef,
) (
	r0 NSURL,
) {
	ret := C.NSURL_type_fileURLWithPathComponents_(
		objc.RefPointer(components),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func NSURL_absoluteURLWithDataRepresentation_relativeToURL_(
	data NSDataRef,
	baseURL NSURLRef,
) (
	r0 NSURL,
) {
	ret := C.NSURL_type_absoluteURLWithDataRepresentation_relativeToURL_(
		objc.RefPointer(data),
		objc.RefPointer(baseURL),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func NSURL_URLWithDataRepresentation_relativeToURL_(
	data NSDataRef,
	baseURL NSURLRef,
) (
	r0 NSURL,
) {
	ret := C.NSURL_type_URLWithDataRepresentation_relativeToURL_(
		objc.RefPointer(data),
		objc.RefPointer(baseURL),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func NSURL_bookmarkDataWithContentsOfURL_error_(
	bookmarkFileURL NSURLRef,
	error NSErrorRef,
) (
	r0 NSData,
) {
	ret := C.NSURL_type_bookmarkDataWithContentsOfURL_error_(
		objc.RefPointer(bookmarkFileURL),
		objc.RefPointer(error),
	)
	r0 = NSData_fromPointer(ret)
	return
}

func NSURL_resourceValuesForKeys_fromBookmarkData_(
	keys NSArrayRef,
	bookmarkData NSDataRef,
) (
	r0 NSDictionary,
) {
	ret := C.NSURL_type_resourceValuesForKeys_fromBookmarkData_(
		objc.RefPointer(keys),
		objc.RefPointer(bookmarkData),
	)
	r0 = NSDictionary_fromPointer(ret)
	return
}

func NSURLRequest_alloc() (
	r0 NSURLRequest,
) {
	ret := C.NSURLRequest_type_alloc()
	r0 = NSURLRequest_fromPointer(ret)
	return
}

func NSURLRequest_requestWithURL_(
	URL NSURLRef,
) (
	r0 NSURLRequest,
) {
	ret := C.NSURLRequest_type_requestWithURL_(
		objc.RefPointer(URL),
	)
	r0 = NSURLRequest_fromPointer(ret)
	return
}

func NSURLRequest_supportsSecureCoding() (
	r0 bool,
) {
	ret := C.NSURLRequest_type_supportsSecureCoding()
	r0 = convertObjCBoolToGo(ret)
	return
}

func NSNotification_alloc() (
	r0 NSNotification,
) {
	ret := C.NSNotification_type_alloc()
	r0 = NSNotification_fromPointer(ret)
	return
}

func NSOperationQueue_alloc() (
	r0 NSOperationQueue,
) {
	ret := C.NSOperationQueue_type_alloc()
	r0 = NSOperationQueue_fromPointer(ret)
	return
}

func NSOperationQueue_mainQueue() (
	r0 NSOperationQueue,
) {
	ret := C.NSOperationQueue_type_mainQueue()
	r0 = NSOperationQueue_fromPointer(ret)
	return
}

func NSOperationQueue_currentQueue() (
	r0 NSOperationQueue,
) {
	ret := C.NSOperationQueue_type_currentQueue()
	r0 = NSOperationQueue_fromPointer(ret)
	return
}

func NSNotificationCenter_alloc() (
	r0 NSNotificationCenter,
) {
	ret := C.NSNotificationCenter_type_alloc()
	r0 = NSNotificationCenter_fromPointer(ret)
	return
}

func NSNotificationCenter_defaultCenter() (
	r0 NSNotificationCenter,
) {
	ret := C.NSNotificationCenter_type_defaultCenter()
	r0 = NSNotificationCenter_fromPointer(ret)
	return
}

type CALayerRef interface {
	Pointer() uintptr
	Init_asCALayer() CALayer
}

type gen_CALayer struct {
	objc.Object
}

func CALayer_fromPointer(ptr unsafe.Pointer) CALayer {
	return CALayer{gen_CALayer{
		objc.Object_fromPointer(ptr),
	}}
}

func CALayer_fromRef(ref objc.Ref) CALayer {
	return CALayer_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_CALayer) Init_asCALayer() (
	r0 CALayer,
) {
	ret := C.CALayer_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = CALayer_fromPointer(ret)
	return
}

func (x gen_CALayer) InitWithLayer__asCALayer(
	layer objc.Ref,
) (
	r0 CALayer,
) {
	ret := C.CALayer_inst_initWithLayer_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(layer),
	)
	r0 = CALayer_fromPointer(ret)
	return
}

func (x gen_CALayer) PresentationLayer_asCALayer() (
	r0 CALayer,
) {
	ret := C.CALayer_inst_presentationLayer(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = CALayer_fromPointer(ret)
	return
}

func (x gen_CALayer) ModelLayer_asCALayer() (
	r0 CALayer,
) {
	ret := C.CALayer_inst_modelLayer(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = CALayer_fromPointer(ret)
	return
}

func (x gen_CALayer) Display() {
	C.CALayer_inst_display(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_CALayer) ContentsAreFlipped() (
	r0 bool,
) {
	ret := C.CALayer_inst_contentsAreFlipped(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_CALayer) AddSublayer_(
	layer CALayerRef,
) {
	C.CALayer_inst_addSublayer_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(layer),
	)
	return
}

func (x gen_CALayer) RemoveFromSuperlayer() {
	C.CALayer_inst_removeFromSuperlayer(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_CALayer) InsertSublayer_atIndex_(
	layer CALayerRef,
	idx int32,
) {
	C.CALayer_inst_insertSublayer_atIndex_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(layer),
		C.int(idx),
	)
	return
}

func (x gen_CALayer) InsertSublayer_below_(
	layer CALayerRef,
	sibling CALayerRef,
) {
	C.CALayer_inst_insertSublayer_below_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(layer),
		objc.RefPointer(sibling),
	)
	return
}

func (x gen_CALayer) InsertSublayer_above_(
	layer CALayerRef,
	sibling CALayerRef,
) {
	C.CALayer_inst_insertSublayer_above_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(layer),
		objc.RefPointer(sibling),
	)
	return
}

func (x gen_CALayer) ReplaceSublayer_with_(
	oldLayer CALayerRef,
	newLayer CALayerRef,
) {
	C.CALayer_inst_replaceSublayer_with_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(oldLayer),
		objc.RefPointer(newLayer),
	)
	return
}

func (x gen_CALayer) SetNeedsDisplay() {
	C.CALayer_inst_setNeedsDisplay(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_CALayer) SetNeedsDisplayInRect_(
	r NSRect,
) {
	C.CALayer_inst_setNeedsDisplayInRect_(
		unsafe.Pointer(x.Pointer()),
		*(*C.NSRect)(unsafe.Pointer(&r)),
	)
	return
}

func (x gen_CALayer) DisplayIfNeeded() {
	C.CALayer_inst_displayIfNeeded(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_CALayer) NeedsDisplay() (
	r0 bool,
) {
	ret := C.CALayer_inst_needsDisplay(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_CALayer) RemoveAllAnimations() {
	C.CALayer_inst_removeAllAnimations(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_CALayer) RemoveAnimationForKey_(
	key NSStringRef,
) {
	C.CALayer_inst_removeAnimationForKey_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(key),
	)
	return
}

func (x gen_CALayer) AnimationKeys() (
	r0 NSArray,
) {
	ret := C.CALayer_inst_animationKeys(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_CALayer) SetNeedsLayout() {
	C.CALayer_inst_setNeedsLayout(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_CALayer) LayoutSublayers() {
	C.CALayer_inst_layoutSublayers(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_CALayer) LayoutIfNeeded() {
	C.CALayer_inst_layoutIfNeeded(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_CALayer) NeedsLayout() (
	r0 bool,
) {
	ret := C.CALayer_inst_needsLayout(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_CALayer) ResizeWithOldSuperlayerSize_(
	size NSSize,
) {
	C.CALayer_inst_resizeWithOldSuperlayerSize_(
		unsafe.Pointer(x.Pointer()),
		*(*C.NSSize)(unsafe.Pointer(&size)),
	)
	return
}

func (x gen_CALayer) ResizeSublayersWithOldSize_(
	size NSSize,
) {
	C.CALayer_inst_resizeSublayersWithOldSize_(
		unsafe.Pointer(x.Pointer()),
		*(*C.NSSize)(unsafe.Pointer(&size)),
	)
	return
}

func (x gen_CALayer) PreferredFrameSize() (
	r0 NSSize,
) {
	ret := C.CALayer_inst_preferredFrameSize(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*NSSize)(unsafe.Pointer(&ret))
	return
}

func (x gen_CALayer) ActionForKey_(
	event NSStringRef,
) (
	r0 objc.Object,
) {
	ret := C.CALayer_inst_actionForKey_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(event),
	)
	r0 = objc.Object_fromPointer(ret)
	return
}

func (x gen_CALayer) ConvertRect_fromLayer_(
	r NSRect,
	l CALayerRef,
) (
	r0 NSRect,
) {
	ret := C.CALayer_inst_convertRect_fromLayer_(
		unsafe.Pointer(x.Pointer()),
		*(*C.NSRect)(unsafe.Pointer(&r)),
		objc.RefPointer(l),
	)
	r0 = *(*NSRect)(unsafe.Pointer(&ret))
	return
}

func (x gen_CALayer) ConvertRect_toLayer_(
	r NSRect,
	l CALayerRef,
) (
	r0 NSRect,
) {
	ret := C.CALayer_inst_convertRect_toLayer_(
		unsafe.Pointer(x.Pointer()),
		*(*C.NSRect)(unsafe.Pointer(&r)),
		objc.RefPointer(l),
	)
	r0 = *(*NSRect)(unsafe.Pointer(&ret))
	return
}

func (x gen_CALayer) ScrollRectToVisible_(
	r NSRect,
) {
	C.CALayer_inst_scrollRectToVisible_(
		unsafe.Pointer(x.Pointer()),
		*(*C.NSRect)(unsafe.Pointer(&r)),
	)
	return
}

func (x gen_CALayer) ShouldArchiveValueForKey_(
	key NSStringRef,
) (
	r0 bool,
) {
	ret := C.CALayer_inst_shouldArchiveValueForKey_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(key),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_CALayer) Delegate() (
	r0 objc.Object,
) {
	ret := C.CALayer_inst_delegate(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = objc.Object_fromPointer(ret)
	return
}

func (x gen_CALayer) SetDelegate_(
	value objc.Ref,
) {
	C.CALayer_inst_setDelegate_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

func (x gen_CALayer) Contents() (
	r0 objc.Object,
) {
	ret := C.CALayer_inst_contents(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = objc.Object_fromPointer(ret)
	return
}

func (x gen_CALayer) SetContents_(
	value objc.Ref,
) {
	C.CALayer_inst_setContents_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

func (x gen_CALayer) ContentsRect() (
	r0 NSRect,
) {
	ret := C.CALayer_inst_contentsRect(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*NSRect)(unsafe.Pointer(&ret))
	return
}

func (x gen_CALayer) SetContentsRect_(
	value NSRect,
) {
	C.CALayer_inst_setContentsRect_(
		unsafe.Pointer(x.Pointer()),
		*(*C.NSRect)(unsafe.Pointer(&value)),
	)
	return
}

func (x gen_CALayer) ContentsCenter() (
	r0 NSRect,
) {
	ret := C.CALayer_inst_contentsCenter(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*NSRect)(unsafe.Pointer(&ret))
	return
}

func (x gen_CALayer) SetContentsCenter_(
	value NSRect,
) {
	C.CALayer_inst_setContentsCenter_(
		unsafe.Pointer(x.Pointer()),
		*(*C.NSRect)(unsafe.Pointer(&value)),
	)
	return
}

func (x gen_CALayer) Opacity() (
	r0 float32,
) {
	ret := C.CALayer_inst_opacity(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float32(ret)
	return
}

func (x gen_CALayer) SetOpacity_(
	value float32,
) {
	C.CALayer_inst_setOpacity_(
		unsafe.Pointer(x.Pointer()),
		C.float(value),
	)
	return
}

func (x gen_CALayer) IsHidden() (
	r0 bool,
) {
	ret := C.CALayer_inst_isHidden(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_CALayer) SetHidden_(
	value bool,
) {
	C.CALayer_inst_setHidden_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_CALayer) MasksToBounds() (
	r0 bool,
) {
	ret := C.CALayer_inst_masksToBounds(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_CALayer) SetMasksToBounds_(
	value bool,
) {
	C.CALayer_inst_setMasksToBounds_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_CALayer) Mask() (
	r0 CALayer,
) {
	ret := C.CALayer_inst_mask(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = CALayer_fromPointer(ret)
	return
}

func (x gen_CALayer) SetMask_(
	value CALayerRef,
) {
	C.CALayer_inst_setMask_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

func (x gen_CALayer) IsDoubleSided() (
	r0 bool,
) {
	ret := C.CALayer_inst_isDoubleSided(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_CALayer) SetDoubleSided_(
	value bool,
) {
	C.CALayer_inst_setDoubleSided_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_CALayer) CornerRadius() (
	r0 CGFloat,
) {
	ret := C.CALayer_inst_cornerRadius(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = CGFloat(ret)
	return
}

func (x gen_CALayer) SetCornerRadius_(
	value CGFloat,
) {
	C.CALayer_inst_setCornerRadius_(
		unsafe.Pointer(x.Pointer()),
		C.double(value),
	)
	return
}

func (x gen_CALayer) BorderWidth() (
	r0 CGFloat,
) {
	ret := C.CALayer_inst_borderWidth(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = CGFloat(ret)
	return
}

func (x gen_CALayer) SetBorderWidth_(
	value CGFloat,
) {
	C.CALayer_inst_setBorderWidth_(
		unsafe.Pointer(x.Pointer()),
		C.double(value),
	)
	return
}

func (x gen_CALayer) ShadowOpacity() (
	r0 float32,
) {
	ret := C.CALayer_inst_shadowOpacity(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float32(ret)
	return
}

func (x gen_CALayer) SetShadowOpacity_(
	value float32,
) {
	C.CALayer_inst_setShadowOpacity_(
		unsafe.Pointer(x.Pointer()),
		C.float(value),
	)
	return
}

func (x gen_CALayer) ShadowRadius() (
	r0 CGFloat,
) {
	ret := C.CALayer_inst_shadowRadius(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = CGFloat(ret)
	return
}

func (x gen_CALayer) SetShadowRadius_(
	value CGFloat,
) {
	C.CALayer_inst_setShadowRadius_(
		unsafe.Pointer(x.Pointer()),
		C.double(value),
	)
	return
}

func (x gen_CALayer) ShadowOffset() (
	r0 NSSize,
) {
	ret := C.CALayer_inst_shadowOffset(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*NSSize)(unsafe.Pointer(&ret))
	return
}

func (x gen_CALayer) SetShadowOffset_(
	value NSSize,
) {
	C.CALayer_inst_setShadowOffset_(
		unsafe.Pointer(x.Pointer()),
		*(*C.NSSize)(unsafe.Pointer(&value)),
	)
	return
}

func (x gen_CALayer) Style() (
	r0 NSDictionary,
) {
	ret := C.CALayer_inst_style(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSDictionary_fromPointer(ret)
	return
}

func (x gen_CALayer) SetStyle_(
	value NSDictionaryRef,
) {
	C.CALayer_inst_setStyle_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

func (x gen_CALayer) AllowsEdgeAntialiasing() (
	r0 bool,
) {
	ret := C.CALayer_inst_allowsEdgeAntialiasing(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_CALayer) SetAllowsEdgeAntialiasing_(
	value bool,
) {
	C.CALayer_inst_setAllowsEdgeAntialiasing_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_CALayer) AllowsGroupOpacity() (
	r0 bool,
) {
	ret := C.CALayer_inst_allowsGroupOpacity(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_CALayer) SetAllowsGroupOpacity_(
	value bool,
) {
	C.CALayer_inst_setAllowsGroupOpacity_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_CALayer) Filters() (
	r0 NSArray,
) {
	ret := C.CALayer_inst_filters(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_CALayer) SetFilters_(
	value NSArrayRef,
) {
	C.CALayer_inst_setFilters_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

func (x gen_CALayer) CompositingFilter() (
	r0 objc.Object,
) {
	ret := C.CALayer_inst_compositingFilter(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = objc.Object_fromPointer(ret)
	return
}

func (x gen_CALayer) SetCompositingFilter_(
	value objc.Ref,
) {
	C.CALayer_inst_setCompositingFilter_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

func (x gen_CALayer) BackgroundFilters() (
	r0 NSArray,
) {
	ret := C.CALayer_inst_backgroundFilters(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_CALayer) SetBackgroundFilters_(
	value NSArrayRef,
) {
	C.CALayer_inst_setBackgroundFilters_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

func (x gen_CALayer) MinificationFilterBias() (
	r0 float32,
) {
	ret := C.CALayer_inst_minificationFilterBias(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float32(ret)
	return
}

func (x gen_CALayer) SetMinificationFilterBias_(
	value float32,
) {
	C.CALayer_inst_setMinificationFilterBias_(
		unsafe.Pointer(x.Pointer()),
		C.float(value),
	)
	return
}

func (x gen_CALayer) IsOpaque() (
	r0 bool,
) {
	ret := C.CALayer_inst_isOpaque(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_CALayer) SetOpaque_(
	value bool,
) {
	C.CALayer_inst_setOpaque_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_CALayer) IsGeometryFlipped() (
	r0 bool,
) {
	ret := C.CALayer_inst_isGeometryFlipped(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_CALayer) SetGeometryFlipped_(
	value bool,
) {
	C.CALayer_inst_setGeometryFlipped_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_CALayer) DrawsAsynchronously() (
	r0 bool,
) {
	ret := C.CALayer_inst_drawsAsynchronously(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_CALayer) SetDrawsAsynchronously_(
	value bool,
) {
	C.CALayer_inst_setDrawsAsynchronously_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_CALayer) ShouldRasterize() (
	r0 bool,
) {
	ret := C.CALayer_inst_shouldRasterize(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_CALayer) SetShouldRasterize_(
	value bool,
) {
	C.CALayer_inst_setShouldRasterize_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_CALayer) RasterizationScale() (
	r0 CGFloat,
) {
	ret := C.CALayer_inst_rasterizationScale(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = CGFloat(ret)
	return
}

func (x gen_CALayer) SetRasterizationScale_(
	value CGFloat,
) {
	C.CALayer_inst_setRasterizationScale_(
		unsafe.Pointer(x.Pointer()),
		C.double(value),
	)
	return
}

func (x gen_CALayer) Frame() (
	r0 NSRect,
) {
	ret := C.CALayer_inst_frame(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*NSRect)(unsafe.Pointer(&ret))
	return
}

func (x gen_CALayer) SetFrame_(
	value NSRect,
) {
	C.CALayer_inst_setFrame_(
		unsafe.Pointer(x.Pointer()),
		*(*C.NSRect)(unsafe.Pointer(&value)),
	)
	return
}

func (x gen_CALayer) Bounds() (
	r0 NSRect,
) {
	ret := C.CALayer_inst_bounds(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*NSRect)(unsafe.Pointer(&ret))
	return
}

func (x gen_CALayer) SetBounds_(
	value NSRect,
) {
	C.CALayer_inst_setBounds_(
		unsafe.Pointer(x.Pointer()),
		*(*C.NSRect)(unsafe.Pointer(&value)),
	)
	return
}

func (x gen_CALayer) ZPosition() (
	r0 CGFloat,
) {
	ret := C.CALayer_inst_zPosition(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = CGFloat(ret)
	return
}

func (x gen_CALayer) SetZPosition_(
	value CGFloat,
) {
	C.CALayer_inst_setZPosition_(
		unsafe.Pointer(x.Pointer()),
		C.double(value),
	)
	return
}

func (x gen_CALayer) AnchorPointZ() (
	r0 CGFloat,
) {
	ret := C.CALayer_inst_anchorPointZ(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = CGFloat(ret)
	return
}

func (x gen_CALayer) SetAnchorPointZ_(
	value CGFloat,
) {
	C.CALayer_inst_setAnchorPointZ_(
		unsafe.Pointer(x.Pointer()),
		C.double(value),
	)
	return
}

func (x gen_CALayer) ContentsScale() (
	r0 CGFloat,
) {
	ret := C.CALayer_inst_contentsScale(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = CGFloat(ret)
	return
}

func (x gen_CALayer) SetContentsScale_(
	value CGFloat,
) {
	C.CALayer_inst_setContentsScale_(
		unsafe.Pointer(x.Pointer()),
		C.double(value),
	)
	return
}

func (x gen_CALayer) Sublayers() (
	r0 NSArray,
) {
	ret := C.CALayer_inst_sublayers(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_CALayer) SetSublayers_(
	value NSArrayRef,
) {
	C.CALayer_inst_setSublayers_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

func (x gen_CALayer) Superlayer() (
	r0 CALayer,
) {
	ret := C.CALayer_inst_superlayer(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = CALayer_fromPointer(ret)
	return
}

func (x gen_CALayer) NeedsDisplayOnBoundsChange() (
	r0 bool,
) {
	ret := C.CALayer_inst_needsDisplayOnBoundsChange(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_CALayer) SetNeedsDisplayOnBoundsChange_(
	value bool,
) {
	C.CALayer_inst_setNeedsDisplayOnBoundsChange_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_CALayer) LayoutManager() (
	r0 objc.Object,
) {
	ret := C.CALayer_inst_layoutManager(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = objc.Object_fromPointer(ret)
	return
}

func (x gen_CALayer) SetLayoutManager_(
	value objc.Ref,
) {
	C.CALayer_inst_setLayoutManager_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

func (x gen_CALayer) Constraints() (
	r0 NSArray,
) {
	ret := C.CALayer_inst_constraints(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_CALayer) SetConstraints_(
	value NSArrayRef,
) {
	C.CALayer_inst_setConstraints_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

func (x gen_CALayer) Actions() (
	r0 NSDictionary,
) {
	ret := C.CALayer_inst_actions(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSDictionary_fromPointer(ret)
	return
}

func (x gen_CALayer) SetActions_(
	value NSDictionaryRef,
) {
	C.CALayer_inst_setActions_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

func (x gen_CALayer) VisibleRect() (
	r0 NSRect,
) {
	ret := C.CALayer_inst_visibleRect(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*NSRect)(unsafe.Pointer(&ret))
	return
}

func (x gen_CALayer) Name() (
	r0 NSString,
) {
	ret := C.CALayer_inst_name(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_CALayer) SetName_(
	value NSStringRef,
) {
	C.CALayer_inst_setName_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

type NSArrayRef interface {
	Pointer() uintptr
	Init_asNSArray() NSArray
}

type gen_NSArray struct {
	objc.Object
}

func NSArray_fromPointer(ptr unsafe.Pointer) NSArray {
	return NSArray{gen_NSArray{
		objc.Object_fromPointer(ptr),
	}}
}

func NSArray_fromRef(ref objc.Ref) NSArray {
	return NSArray_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_NSArray) Init_asNSArray() (
	r0 NSArray,
) {
	ret := C.NSArray_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSArray) InitWithArray__asNSArray(
	array NSArrayRef,
) (
	r0 NSArray,
) {
	ret := C.NSArray_inst_initWithArray_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(array),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSArray) InitWithArray_copyItems__asNSArray(
	array NSArrayRef,
	flag bool,
) (
	r0 NSArray,
) {
	ret := C.NSArray_inst_initWithArray_copyItems_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(array),
		convertToObjCBool(flag),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSArray) ContainsObject_(
	anObject objc.Ref,
) (
	r0 bool,
) {
	ret := C.NSArray_inst_containsObject_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(anObject),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSArray) ObjectAtIndex_(
	index NSUInteger,
) (
	r0 objc.Object,
) {
	ret := C.NSArray_inst_objectAtIndex_(
		unsafe.Pointer(x.Pointer()),
		C.ulong(index),
	)
	r0 = objc.Object_fromPointer(ret)
	return
}

func (x gen_NSArray) ObjectAtIndexedSubscript_(
	idx NSUInteger,
) (
	r0 objc.Object,
) {
	ret := C.NSArray_inst_objectAtIndexedSubscript_(
		unsafe.Pointer(x.Pointer()),
		C.ulong(idx),
	)
	r0 = objc.Object_fromPointer(ret)
	return
}

func (x gen_NSArray) IndexOfObject_(
	anObject objc.Ref,
) (
	r0 NSUInteger,
) {
	ret := C.NSArray_inst_indexOfObject_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(anObject),
	)
	r0 = NSUInteger(ret)
	return
}

func (x gen_NSArray) IndexOfObjectIdenticalTo_(
	anObject objc.Ref,
) (
	r0 NSUInteger,
) {
	ret := C.NSArray_inst_indexOfObjectIdenticalTo_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(anObject),
	)
	r0 = NSUInteger(ret)
	return
}

func (x gen_NSArray) MakeObjectsPerformSelector_(
	aSelector objc.Selector,
) {
	C.NSArray_inst_makeObjectsPerformSelector_(
		unsafe.Pointer(x.Pointer()),
		aSelector.SelectorAddress(),
	)
	return
}

func (x gen_NSArray) MakeObjectsPerformSelector_withObject_(
	aSelector objc.Selector,
	argument objc.Ref,
) {
	C.NSArray_inst_makeObjectsPerformSelector_withObject_(
		unsafe.Pointer(x.Pointer()),
		aSelector.SelectorAddress(),
		objc.RefPointer(argument),
	)
	return
}

func (x gen_NSArray) FirstObjectCommonWithArray_(
	otherArray NSArrayRef,
) (
	r0 objc.Object,
) {
	ret := C.NSArray_inst_firstObjectCommonWithArray_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(otherArray),
	)
	r0 = objc.Object_fromPointer(ret)
	return
}

func (x gen_NSArray) IsEqualToArray_(
	otherArray NSArrayRef,
) (
	r0 bool,
) {
	ret := C.NSArray_inst_isEqualToArray_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(otherArray),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSArray) ArrayByAddingObject_(
	anObject objc.Ref,
) (
	r0 NSArray,
) {
	ret := C.NSArray_inst_arrayByAddingObject_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(anObject),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSArray) ArrayByAddingObjectsFromArray_(
	otherArray NSArrayRef,
) (
	r0 NSArray,
) {
	ret := C.NSArray_inst_arrayByAddingObjectsFromArray_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(otherArray),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSArray) SortedArrayUsingDescriptors_(
	sortDescriptors NSArrayRef,
) (
	r0 NSArray,
) {
	ret := C.NSArray_inst_sortedArrayUsingDescriptors_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(sortDescriptors),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSArray) SortedArrayUsingSelector_(
	comparator objc.Selector,
) (
	r0 NSArray,
) {
	ret := C.NSArray_inst_sortedArrayUsingSelector_(
		unsafe.Pointer(x.Pointer()),
		comparator.SelectorAddress(),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSArray) ComponentsJoinedByString_(
	separator NSStringRef,
) (
	r0 NSString,
) {
	ret := C.NSArray_inst_componentsJoinedByString_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(separator),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSArray) DescriptionWithLocale_(
	locale objc.Ref,
) (
	r0 NSString,
) {
	ret := C.NSArray_inst_descriptionWithLocale_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(locale),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSArray) DescriptionWithLocale_indent_(
	locale objc.Ref,
	level NSUInteger,
) (
	r0 NSString,
) {
	ret := C.NSArray_inst_descriptionWithLocale_indent_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(locale),
		C.ulong(level),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSArray) PathsMatchingExtensions_(
	filterTypes NSArrayRef,
) (
	r0 NSArray,
) {
	ret := C.NSArray_inst_pathsMatchingExtensions_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(filterTypes),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSArray) SetValue_forKey_(
	value objc.Ref,
	key NSStringRef,
) {
	C.NSArray_inst_setValue_forKey_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
		objc.RefPointer(key),
	)
	return
}

func (x gen_NSArray) ValueForKey_(
	key NSStringRef,
) (
	r0 objc.Object,
) {
	ret := C.NSArray_inst_valueForKey_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(key),
	)
	r0 = objc.Object_fromPointer(ret)
	return
}

func (x gen_NSArray) ShuffledArray() (
	r0 NSArray,
) {
	ret := C.NSArray_inst_shuffledArray(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSArray) InitWithContentsOfURL_error_(
	url NSURLRef,
	error NSErrorRef,
) (
	r0 NSArray,
) {
	ret := C.NSArray_inst_initWithContentsOfURL_error_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(url),
		objc.RefPointer(error),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSArray) WriteToURL_error_(
	url NSURLRef,
	error NSErrorRef,
) (
	r0 bool,
) {
	ret := C.NSArray_inst_writeToURL_error_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(url),
		objc.RefPointer(error),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSArray) Count() (
	r0 NSUInteger,
) {
	ret := C.NSArray_inst_count(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSUInteger(ret)
	return
}

func (x gen_NSArray) FirstObject() (
	r0 objc.Object,
) {
	ret := C.NSArray_inst_firstObject(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = objc.Object_fromPointer(ret)
	return
}

func (x gen_NSArray) LastObject() (
	r0 objc.Object,
) {
	ret := C.NSArray_inst_lastObject(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = objc.Object_fromPointer(ret)
	return
}

func (x gen_NSArray) SortedArrayHint() (
	r0 NSData,
) {
	ret := C.NSArray_inst_sortedArrayHint(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSData_fromPointer(ret)
	return
}

func (x gen_NSArray) Description() (
	r0 NSString,
) {
	ret := C.NSArray_inst_description(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

type NSAttributedStringRef interface {
	Pointer() uintptr
	Init_asNSAttributedString() NSAttributedString
}

type gen_NSAttributedString struct {
	objc.Object
}

func NSAttributedString_fromPointer(ptr unsafe.Pointer) NSAttributedString {
	return NSAttributedString{gen_NSAttributedString{
		objc.Object_fromPointer(ptr),
	}}
}

func NSAttributedString_fromRef(ref objc.Ref) NSAttributedString {
	return NSAttributedString_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_NSAttributedString) InitWithString__asNSAttributedString(
	str NSStringRef,
) (
	r0 NSAttributedString,
) {
	ret := C.NSAttributedString_inst_initWithString_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(str),
	)
	r0 = NSAttributedString_fromPointer(ret)
	return
}

func (x gen_NSAttributedString) InitWithString_attributes__asNSAttributedString(
	str NSStringRef,
	attrs NSDictionaryRef,
) (
	r0 NSAttributedString,
) {
	ret := C.NSAttributedString_inst_initWithString_attributes_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(str),
		objc.RefPointer(attrs),
	)
	r0 = NSAttributedString_fromPointer(ret)
	return
}

func (x gen_NSAttributedString) InitWithAttributedString__asNSAttributedString(
	attrStr NSAttributedStringRef,
) (
	r0 NSAttributedString,
) {
	ret := C.NSAttributedString_inst_initWithAttributedString_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(attrStr),
	)
	r0 = NSAttributedString_fromPointer(ret)
	return
}

func (x gen_NSAttributedString) InitWithData_options_documentAttributes_error__asNSAttributedString(
	data NSDataRef,
	options NSDictionaryRef,
	dict NSDictionaryRef,
	error NSErrorRef,
) (
	r0 NSAttributedString,
) {
	ret := C.NSAttributedString_inst_initWithData_options_documentAttributes_error_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(data),
		objc.RefPointer(options),
		objc.RefPointer(dict),
		objc.RefPointer(error),
	)
	r0 = NSAttributedString_fromPointer(ret)
	return
}

func (x gen_NSAttributedString) InitWithDocFormat_documentAttributes__asNSAttributedString(
	data NSDataRef,
	dict NSDictionaryRef,
) (
	r0 NSAttributedString,
) {
	ret := C.NSAttributedString_inst_initWithDocFormat_documentAttributes_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(data),
		objc.RefPointer(dict),
	)
	r0 = NSAttributedString_fromPointer(ret)
	return
}

func (x gen_NSAttributedString) InitWithHTML_documentAttributes__asNSAttributedString(
	data NSDataRef,
	dict NSDictionaryRef,
) (
	r0 NSAttributedString,
) {
	ret := C.NSAttributedString_inst_initWithHTML_documentAttributes_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(data),
		objc.RefPointer(dict),
	)
	r0 = NSAttributedString_fromPointer(ret)
	return
}

func (x gen_NSAttributedString) InitWithHTML_baseURL_documentAttributes__asNSAttributedString(
	data NSDataRef,
	base NSURLRef,
	dict NSDictionaryRef,
) (
	r0 NSAttributedString,
) {
	ret := C.NSAttributedString_inst_initWithHTML_baseURL_documentAttributes_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(data),
		objc.RefPointer(base),
		objc.RefPointer(dict),
	)
	r0 = NSAttributedString_fromPointer(ret)
	return
}

func (x gen_NSAttributedString) InitWithHTML_options_documentAttributes__asNSAttributedString(
	data NSDataRef,
	options NSDictionaryRef,
	dict NSDictionaryRef,
) (
	r0 NSAttributedString,
) {
	ret := C.NSAttributedString_inst_initWithHTML_options_documentAttributes_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(data),
		objc.RefPointer(options),
		objc.RefPointer(dict),
	)
	r0 = NSAttributedString_fromPointer(ret)
	return
}

func (x gen_NSAttributedString) InitWithRTF_documentAttributes__asNSAttributedString(
	data NSDataRef,
	dict NSDictionaryRef,
) (
	r0 NSAttributedString,
) {
	ret := C.NSAttributedString_inst_initWithRTF_documentAttributes_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(data),
		objc.RefPointer(dict),
	)
	r0 = NSAttributedString_fromPointer(ret)
	return
}

func (x gen_NSAttributedString) InitWithRTFD_documentAttributes__asNSAttributedString(
	data NSDataRef,
	dict NSDictionaryRef,
) (
	r0 NSAttributedString,
) {
	ret := C.NSAttributedString_inst_initWithRTFD_documentAttributes_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(data),
		objc.RefPointer(dict),
	)
	r0 = NSAttributedString_fromPointer(ret)
	return
}

func (x gen_NSAttributedString) InitWithURL_options_documentAttributes_error__asNSAttributedString(
	url NSURLRef,
	options NSDictionaryRef,
	dict NSDictionaryRef,
	error NSErrorRef,
) (
	r0 NSAttributedString,
) {
	ret := C.NSAttributedString_inst_initWithURL_options_documentAttributes_error_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(url),
		objc.RefPointer(options),
		objc.RefPointer(dict),
		objc.RefPointer(error),
	)
	r0 = NSAttributedString_fromPointer(ret)
	return
}

func (x gen_NSAttributedString) IsEqualToAttributedString_(
	other NSAttributedStringRef,
) (
	r0 bool,
) {
	ret := C.NSAttributedString_inst_isEqualToAttributedString_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(other),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSAttributedString) NextWordFromIndex_forward_(
	location NSUInteger,
	isForward bool,
) (
	r0 NSUInteger,
) {
	ret := C.NSAttributedString_inst_nextWordFromIndex_forward_(
		unsafe.Pointer(x.Pointer()),
		C.ulong(location),
		convertToObjCBool(isForward),
	)
	r0 = NSUInteger(ret)
	return
}

func (x gen_NSAttributedString) DrawInRect_(
	rect NSRect,
) {
	C.NSAttributedString_inst_drawInRect_(
		unsafe.Pointer(x.Pointer()),
		*(*C.NSRect)(unsafe.Pointer(&rect)),
	)
	return
}

func (x gen_NSAttributedString) Size() (
	r0 NSSize,
) {
	ret := C.NSAttributedString_inst_size(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = *(*NSSize)(unsafe.Pointer(&ret))
	return
}

func (x gen_NSAttributedString) Init_asNSAttributedString() (
	r0 NSAttributedString,
) {
	ret := C.NSAttributedString_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSAttributedString_fromPointer(ret)
	return
}

func (x gen_NSAttributedString) String() (
	r0 NSString,
) {
	ret := C.NSAttributedString_inst_string(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSAttributedString) Length() (
	r0 NSUInteger,
) {
	ret := C.NSAttributedString_inst_length(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSUInteger(ret)
	return
}

type NSDataRef interface {
	Pointer() uintptr
	Init_asNSData() NSData
}

type gen_NSData struct {
	objc.Object
}

func NSData_fromPointer(ptr unsafe.Pointer) NSData {
	return NSData{gen_NSData{
		objc.Object_fromPointer(ptr),
	}}
}

func NSData_fromRef(ref objc.Ref) NSData {
	return NSData_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_NSData) InitWithBytes_length__asNSData(
	bytes unsafe.Pointer,
	length NSUInteger,
) (
	r0 NSData,
) {
	ret := C.NSData_inst_initWithBytes_length_(
		unsafe.Pointer(x.Pointer()),
		bytes,
		C.ulong(length),
	)
	r0 = NSData_fromPointer(ret)
	return
}

func (x gen_NSData) InitWithBytesNoCopy_length__asNSData(
	bytes unsafe.Pointer,
	length NSUInteger,
) (
	r0 NSData,
) {
	ret := C.NSData_inst_initWithBytesNoCopy_length_(
		unsafe.Pointer(x.Pointer()),
		bytes,
		C.ulong(length),
	)
	r0 = NSData_fromPointer(ret)
	return
}

func (x gen_NSData) InitWithBytesNoCopy_length_freeWhenDone__asNSData(
	bytes unsafe.Pointer,
	length NSUInteger,
	b bool,
) (
	r0 NSData,
) {
	ret := C.NSData_inst_initWithBytesNoCopy_length_freeWhenDone_(
		unsafe.Pointer(x.Pointer()),
		bytes,
		C.ulong(length),
		convertToObjCBool(b),
	)
	r0 = NSData_fromPointer(ret)
	return
}

func (x gen_NSData) InitWithData__asNSData(
	data NSDataRef,
) (
	r0 NSData,
) {
	ret := C.NSData_inst_initWithData_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(data),
	)
	r0 = NSData_fromPointer(ret)
	return
}

func (x gen_NSData) InitWithContentsOfFile__asNSData(
	path NSStringRef,
) (
	r0 NSData,
) {
	ret := C.NSData_inst_initWithContentsOfFile_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(path),
	)
	r0 = NSData_fromPointer(ret)
	return
}

func (x gen_NSData) InitWithContentsOfURL__asNSData(
	url NSURLRef,
) (
	r0 NSData,
) {
	ret := C.NSData_inst_initWithContentsOfURL_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(url),
	)
	r0 = NSData_fromPointer(ret)
	return
}

func (x gen_NSData) WriteToFile_atomically_(
	path NSStringRef,
	useAuxiliaryFile bool,
) (
	r0 bool,
) {
	ret := C.NSData_inst_writeToFile_atomically_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(path),
		convertToObjCBool(useAuxiliaryFile),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSData) WriteToURL_atomically_(
	url NSURLRef,
	atomically bool,
) (
	r0 bool,
) {
	ret := C.NSData_inst_writeToURL_atomically_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(url),
		convertToObjCBool(atomically),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSData) GetBytes_length_(
	buffer unsafe.Pointer,
	length NSUInteger,
) {
	C.NSData_inst_getBytes_length_(
		unsafe.Pointer(x.Pointer()),
		buffer,
		C.ulong(length),
	)
	return
}

func (x gen_NSData) IsEqualToData_(
	other NSDataRef,
) (
	r0 bool,
) {
	ret := C.NSData_inst_isEqualToData_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(other),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSData) Init_asNSData() (
	r0 NSData,
) {
	ret := C.NSData_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSData_fromPointer(ret)
	return
}

func (x gen_NSData) Bytes() (
	r0 unsafe.Pointer,
) {
	ret := C.NSData_inst_bytes(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = ret
	return
}

func (x gen_NSData) Length() (
	r0 NSUInteger,
) {
	ret := C.NSData_inst_length(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSUInteger(ret)
	return
}

func (x gen_NSData) Description() (
	r0 NSString,
) {
	ret := C.NSData_inst_description(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

type NSDateRef interface {
	Pointer() uintptr
	Init_asNSDate() NSDate
}

type gen_NSDate struct {
	objc.Object
}

func NSDate_fromPointer(ptr unsafe.Pointer) NSDate {
	return NSDate{gen_NSDate{
		objc.Object_fromPointer(ptr),
	}}
}

func NSDate_fromRef(ref objc.Ref) NSDate {
	return NSDate_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_NSDate) Init_asNSDate() (
	r0 NSDate,
) {
	ret := C.NSDate_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSDate_fromPointer(ret)
	return
}

func (x gen_NSDate) IsEqualToDate_(
	otherDate NSDateRef,
) (
	r0 bool,
) {
	ret := C.NSDate_inst_isEqualToDate_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(otherDate),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSDate) EarlierDate_(
	anotherDate NSDateRef,
) (
	r0 NSDate,
) {
	ret := C.NSDate_inst_earlierDate_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(anotherDate),
	)
	r0 = NSDate_fromPointer(ret)
	return
}

func (x gen_NSDate) LaterDate_(
	anotherDate NSDateRef,
) (
	r0 NSDate,
) {
	ret := C.NSDate_inst_laterDate_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(anotherDate),
	)
	r0 = NSDate_fromPointer(ret)
	return
}

func (x gen_NSDate) DescriptionWithLocale_(
	locale objc.Ref,
) (
	r0 NSString,
) {
	ret := C.NSDate_inst_descriptionWithLocale_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(locale),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSDate) Description() (
	r0 NSString,
) {
	ret := C.NSDate_inst_description(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

type NSDictionaryRef interface {
	Pointer() uintptr
	Init_asNSDictionary() NSDictionary
}

type gen_NSDictionary struct {
	objc.Object
}

func NSDictionary_fromPointer(ptr unsafe.Pointer) NSDictionary {
	return NSDictionary{gen_NSDictionary{
		objc.Object_fromPointer(ptr),
	}}
}

func NSDictionary_fromRef(ref objc.Ref) NSDictionary {
	return NSDictionary_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_NSDictionary) Init_asNSDictionary() (
	r0 NSDictionary,
) {
	ret := C.NSDictionary_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSDictionary_fromPointer(ret)
	return
}

func (x gen_NSDictionary) InitWithObjects_forKeys__asNSDictionary(
	objects NSArrayRef,
	keys NSArrayRef,
) (
	r0 NSDictionary,
) {
	ret := C.NSDictionary_inst_initWithObjects_forKeys_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(objects),
		objc.RefPointer(keys),
	)
	r0 = NSDictionary_fromPointer(ret)
	return
}

func (x gen_NSDictionary) InitWithDictionary__asNSDictionary(
	otherDictionary NSDictionaryRef,
) (
	r0 NSDictionary,
) {
	ret := C.NSDictionary_inst_initWithDictionary_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(otherDictionary),
	)
	r0 = NSDictionary_fromPointer(ret)
	return
}

func (x gen_NSDictionary) InitWithDictionary_copyItems__asNSDictionary(
	otherDictionary NSDictionaryRef,
	flag bool,
) (
	r0 NSDictionary,
) {
	ret := C.NSDictionary_inst_initWithDictionary_copyItems_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(otherDictionary),
		convertToObjCBool(flag),
	)
	r0 = NSDictionary_fromPointer(ret)
	return
}

func (x gen_NSDictionary) InitWithContentsOfURL_error_(
	url NSURLRef,
	error NSErrorRef,
) (
	r0 NSDictionary,
) {
	ret := C.NSDictionary_inst_initWithContentsOfURL_error_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(url),
		objc.RefPointer(error),
	)
	r0 = NSDictionary_fromPointer(ret)
	return
}

func (x gen_NSDictionary) IsEqualToDictionary_(
	otherDictionary NSDictionaryRef,
) (
	r0 bool,
) {
	ret := C.NSDictionary_inst_isEqualToDictionary_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(otherDictionary),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSDictionary) AllKeysForObject_(
	anObject objc.Ref,
) (
	r0 NSArray,
) {
	ret := C.NSDictionary_inst_allKeysForObject_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(anObject),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSDictionary) ValueForKey_(
	key NSStringRef,
) (
	r0 objc.Object,
) {
	ret := C.NSDictionary_inst_valueForKey_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(key),
	)
	r0 = objc.Object_fromPointer(ret)
	return
}

func (x gen_NSDictionary) ObjectsForKeys_notFoundMarker_(
	keys NSArrayRef,
	marker objc.Ref,
) (
	r0 NSArray,
) {
	ret := C.NSDictionary_inst_objectsForKeys_notFoundMarker_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(keys),
		objc.RefPointer(marker),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSDictionary) KeysSortedByValueUsingSelector_(
	comparator objc.Selector,
) (
	r0 NSArray,
) {
	ret := C.NSDictionary_inst_keysSortedByValueUsingSelector_(
		unsafe.Pointer(x.Pointer()),
		comparator.SelectorAddress(),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSDictionary) WriteToURL_error_(
	url NSURLRef,
	error NSErrorRef,
) (
	r0 bool,
) {
	ret := C.NSDictionary_inst_writeToURL_error_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(url),
		objc.RefPointer(error),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSDictionary) FileType() (
	r0 NSString,
) {
	ret := C.NSDictionary_inst_fileType(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSDictionary) FileCreationDate() (
	r0 NSDate,
) {
	ret := C.NSDictionary_inst_fileCreationDate(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSDate_fromPointer(ret)
	return
}

func (x gen_NSDictionary) FileModificationDate() (
	r0 NSDate,
) {
	ret := C.NSDictionary_inst_fileModificationDate(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSDate_fromPointer(ret)
	return
}

func (x gen_NSDictionary) FilePosixPermissions() (
	r0 NSUInteger,
) {
	ret := C.NSDictionary_inst_filePosixPermissions(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSUInteger(ret)
	return
}

func (x gen_NSDictionary) FileOwnerAccountID() (
	r0 NSNumber,
) {
	ret := C.NSDictionary_inst_fileOwnerAccountID(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSNumber_fromPointer(ret)
	return
}

func (x gen_NSDictionary) FileOwnerAccountName() (
	r0 NSString,
) {
	ret := C.NSDictionary_inst_fileOwnerAccountName(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSDictionary) FileGroupOwnerAccountID() (
	r0 NSNumber,
) {
	ret := C.NSDictionary_inst_fileGroupOwnerAccountID(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSNumber_fromPointer(ret)
	return
}

func (x gen_NSDictionary) FileGroupOwnerAccountName() (
	r0 NSString,
) {
	ret := C.NSDictionary_inst_fileGroupOwnerAccountName(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSDictionary) FileExtensionHidden() (
	r0 bool,
) {
	ret := C.NSDictionary_inst_fileExtensionHidden(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSDictionary) FileIsImmutable() (
	r0 bool,
) {
	ret := C.NSDictionary_inst_fileIsImmutable(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSDictionary) FileIsAppendOnly() (
	r0 bool,
) {
	ret := C.NSDictionary_inst_fileIsAppendOnly(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSDictionary) FileSystemFileNumber() (
	r0 NSUInteger,
) {
	ret := C.NSDictionary_inst_fileSystemFileNumber(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSUInteger(ret)
	return
}

func (x gen_NSDictionary) FileSystemNumber() (
	r0 NSInteger,
) {
	ret := C.NSDictionary_inst_fileSystemNumber(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSInteger(ret)
	return
}

func (x gen_NSDictionary) DescriptionWithLocale_(
	locale objc.Ref,
) (
	r0 NSString,
) {
	ret := C.NSDictionary_inst_descriptionWithLocale_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(locale),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSDictionary) DescriptionWithLocale_indent_(
	locale objc.Ref,
	level NSUInteger,
) (
	r0 NSString,
) {
	ret := C.NSDictionary_inst_descriptionWithLocale_indent_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(locale),
		C.ulong(level),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSDictionary) Count() (
	r0 NSUInteger,
) {
	ret := C.NSDictionary_inst_count(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSUInteger(ret)
	return
}

func (x gen_NSDictionary) AllKeys() (
	r0 NSArray,
) {
	ret := C.NSDictionary_inst_allKeys(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSDictionary) AllValues() (
	r0 NSArray,
) {
	ret := C.NSDictionary_inst_allValues(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSDictionary) Description() (
	r0 NSString,
) {
	ret := C.NSDictionary_inst_description(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSDictionary) DescriptionInStringsFileFormat() (
	r0 NSString,
) {
	ret := C.NSDictionary_inst_descriptionInStringsFileFormat(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

type NSNumberRef interface {
	Pointer() uintptr
	Init_asNSNumber() NSNumber
}

type gen_NSNumber struct {
	objc.Object
}

func NSNumber_fromPointer(ptr unsafe.Pointer) NSNumber {
	return NSNumber{gen_NSNumber{
		objc.Object_fromPointer(ptr),
	}}
}

func NSNumber_fromRef(ref objc.Ref) NSNumber {
	return NSNumber_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_NSNumber) InitWithBool_(
	value bool,
) (
	r0 NSNumber,
) {
	ret := C.NSNumber_inst_initWithBool_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	r0 = NSNumber_fromPointer(ret)
	return
}

func (x gen_NSNumber) InitWithDouble_(
	value float64,
) (
	r0 NSNumber,
) {
	ret := C.NSNumber_inst_initWithDouble_(
		unsafe.Pointer(x.Pointer()),
		C.double(value),
	)
	r0 = NSNumber_fromPointer(ret)
	return
}

func (x gen_NSNumber) InitWithFloat_(
	value float32,
) (
	r0 NSNumber,
) {
	ret := C.NSNumber_inst_initWithFloat_(
		unsafe.Pointer(x.Pointer()),
		C.float(value),
	)
	r0 = NSNumber_fromPointer(ret)
	return
}

func (x gen_NSNumber) InitWithInt_(
	value int32,
) (
	r0 NSNumber,
) {
	ret := C.NSNumber_inst_initWithInt_(
		unsafe.Pointer(x.Pointer()),
		C.int(value),
	)
	r0 = NSNumber_fromPointer(ret)
	return
}

func (x gen_NSNumber) InitWithInteger_(
	value NSInteger,
) (
	r0 NSNumber,
) {
	ret := C.NSNumber_inst_initWithInteger_(
		unsafe.Pointer(x.Pointer()),
		C.long(value),
	)
	r0 = NSNumber_fromPointer(ret)
	return
}

func (x gen_NSNumber) InitWithUnsignedInt_(
	value int32,
) (
	r0 NSNumber,
) {
	ret := C.NSNumber_inst_initWithUnsignedInt_(
		unsafe.Pointer(x.Pointer()),
		C.int(value),
	)
	r0 = NSNumber_fromPointer(ret)
	return
}

func (x gen_NSNumber) InitWithUnsignedInteger_(
	value NSUInteger,
) (
	r0 NSNumber,
) {
	ret := C.NSNumber_inst_initWithUnsignedInteger_(
		unsafe.Pointer(x.Pointer()),
		C.ulong(value),
	)
	r0 = NSNumber_fromPointer(ret)
	return
}

func (x gen_NSNumber) DescriptionWithLocale_(
	locale objc.Ref,
) (
	r0 NSString,
) {
	ret := C.NSNumber_inst_descriptionWithLocale_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(locale),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSNumber) IsEqualToNumber_(
	number NSNumberRef,
) (
	r0 bool,
) {
	ret := C.NSNumber_inst_isEqualToNumber_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(number),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSNumber) Init_asNSNumber() (
	r0 NSNumber,
) {
	ret := C.NSNumber_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSNumber_fromPointer(ret)
	return
}

func (x gen_NSNumber) BoolValue() (
	r0 bool,
) {
	ret := C.NSNumber_inst_boolValue(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSNumber) DoubleValue() (
	r0 float64,
) {
	ret := C.NSNumber_inst_doubleValue(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float64(ret)
	return
}

func (x gen_NSNumber) FloatValue() (
	r0 float32,
) {
	ret := C.NSNumber_inst_floatValue(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float32(ret)
	return
}

func (x gen_NSNumber) IntValue() (
	r0 int32,
) {
	ret := C.NSNumber_inst_intValue(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = int32(ret)
	return
}

func (x gen_NSNumber) IntegerValue() (
	r0 NSInteger,
) {
	ret := C.NSNumber_inst_integerValue(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSInteger(ret)
	return
}

func (x gen_NSNumber) UnsignedIntegerValue() (
	r0 NSUInteger,
) {
	ret := C.NSNumber_inst_unsignedIntegerValue(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSUInteger(ret)
	return
}

func (x gen_NSNumber) UnsignedIntValue() (
	r0 int32,
) {
	ret := C.NSNumber_inst_unsignedIntValue(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = int32(ret)
	return
}

func (x gen_NSNumber) StringValue() (
	r0 NSString,
) {
	ret := C.NSNumber_inst_stringValue(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

type NSRunLoopRef interface {
	Pointer() uintptr
	Init_asNSRunLoop() NSRunLoop
}

type gen_NSRunLoop struct {
	objc.Object
}

func NSRunLoop_fromPointer(ptr unsafe.Pointer) NSRunLoop {
	return NSRunLoop{gen_NSRunLoop{
		objc.Object_fromPointer(ptr),
	}}
}

func NSRunLoop_fromRef(ref objc.Ref) NSRunLoop {
	return NSRunLoop_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_NSRunLoop) Run() {
	C.NSRunLoop_inst_run(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_NSRunLoop) RunUntilDate_(
	limitDate NSDateRef,
) {
	C.NSRunLoop_inst_runUntilDate_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(limitDate),
	)
	return
}

func (x gen_NSRunLoop) PerformSelector_target_argument_order_modes_(
	aSelector objc.Selector,
	target objc.Ref,
	arg objc.Ref,
	order NSUInteger,
	modes NSArrayRef,
) {
	C.NSRunLoop_inst_performSelector_target_argument_order_modes_(
		unsafe.Pointer(x.Pointer()),
		aSelector.SelectorAddress(),
		objc.RefPointer(target),
		objc.RefPointer(arg),
		C.ulong(order),
		objc.RefPointer(modes),
	)
	return
}

func (x gen_NSRunLoop) CancelPerformSelector_target_argument_(
	aSelector objc.Selector,
	target objc.Ref,
	arg objc.Ref,
) {
	C.NSRunLoop_inst_cancelPerformSelector_target_argument_(
		unsafe.Pointer(x.Pointer()),
		aSelector.SelectorAddress(),
		objc.RefPointer(target),
		objc.RefPointer(arg),
	)
	return
}

func (x gen_NSRunLoop) CancelPerformSelectorsWithTarget_(
	target objc.Ref,
) {
	C.NSRunLoop_inst_cancelPerformSelectorsWithTarget_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(target),
	)
	return
}

func (x gen_NSRunLoop) Init_asNSRunLoop() (
	r0 NSRunLoop,
) {
	ret := C.NSRunLoop_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSRunLoop_fromPointer(ret)
	return
}

type NSStringRef interface {
	Pointer() uintptr
	Init_asNSString() NSString
}

type gen_NSString struct {
	objc.Object
}

func NSString_fromPointer(ptr unsafe.Pointer) NSString {
	return NSString{gen_NSString{
		objc.Object_fromPointer(ptr),
	}}
}

func NSString_fromRef(ref objc.Ref) NSString {
	return NSString_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_NSString) Init_asNSString() (
	r0 NSString,
) {
	ret := C.NSString_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) InitWithBytes_length_encoding__asNSString(
	bytes unsafe.Pointer,
	len NSUInteger,
	encoding NSStringEncoding,
) (
	r0 NSString,
) {
	ret := C.NSString_inst_initWithBytes_length_encoding_(
		unsafe.Pointer(x.Pointer()),
		bytes,
		C.ulong(len),
		C.ulong(encoding),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) InitWithBytesNoCopy_length_encoding_freeWhenDone__asNSString(
	bytes unsafe.Pointer,
	len NSUInteger,
	encoding NSStringEncoding,
	freeBuffer bool,
) (
	r0 NSString,
) {
	ret := C.NSString_inst_initWithBytesNoCopy_length_encoding_freeWhenDone_(
		unsafe.Pointer(x.Pointer()),
		bytes,
		C.ulong(len),
		C.ulong(encoding),
		convertToObjCBool(freeBuffer),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) InitWithString__asNSString(
	aString NSStringRef,
) (
	r0 NSString,
) {
	ret := C.NSString_inst_initWithString_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(aString),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) InitWithData_encoding__asNSString(
	data NSDataRef,
	encoding NSStringEncoding,
) (
	r0 NSString,
) {
	ret := C.NSString_inst_initWithData_encoding_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(data),
		C.ulong(encoding),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) InitWithContentsOfFile_encoding_error__asNSString(
	path NSStringRef,
	enc NSStringEncoding,
	error NSErrorRef,
) (
	r0 NSString,
) {
	ret := C.NSString_inst_initWithContentsOfFile_encoding_error_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(path),
		C.ulong(enc),
		objc.RefPointer(error),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) InitWithContentsOfURL_encoding_error__asNSString(
	url NSURLRef,
	enc NSStringEncoding,
	error NSErrorRef,
) (
	r0 NSString,
) {
	ret := C.NSString_inst_initWithContentsOfURL_encoding_error_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(url),
		C.ulong(enc),
		objc.RefPointer(error),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) LengthOfBytesUsingEncoding_(
	enc NSStringEncoding,
) (
	r0 NSUInteger,
) {
	ret := C.NSString_inst_lengthOfBytesUsingEncoding_(
		unsafe.Pointer(x.Pointer()),
		C.ulong(enc),
	)
	r0 = NSUInteger(ret)
	return
}

func (x gen_NSString) MaximumLengthOfBytesUsingEncoding_(
	enc NSStringEncoding,
) (
	r0 NSUInteger,
) {
	ret := C.NSString_inst_maximumLengthOfBytesUsingEncoding_(
		unsafe.Pointer(x.Pointer()),
		C.ulong(enc),
	)
	r0 = NSUInteger(ret)
	return
}

func (x gen_NSString) CharacterAtIndex_(
	index NSUInteger,
) (
	r0 Unichar,
) {
	ret := C.NSString_inst_characterAtIndex_(
		unsafe.Pointer(x.Pointer()),
		C.ulong(index),
	)
	r0 = Unichar(ret)
	return
}

func (x gen_NSString) HasPrefix_(
	str NSStringRef,
) (
	r0 bool,
) {
	ret := C.NSString_inst_hasPrefix_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(str),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSString) HasSuffix_(
	str NSStringRef,
) (
	r0 bool,
) {
	ret := C.NSString_inst_hasSuffix_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(str),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSString) IsEqualToString_(
	aString NSStringRef,
) (
	r0 bool,
) {
	ret := C.NSString_inst_isEqualToString_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(aString),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSString) StringByAppendingString_(
	aString NSStringRef,
) (
	r0 NSString,
) {
	ret := C.NSString_inst_stringByAppendingString_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(aString),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) StringByPaddingToLength_withString_startingAtIndex_(
	newLength NSUInteger,
	padString NSStringRef,
	padIndex NSUInteger,
) (
	r0 NSString,
) {
	ret := C.NSString_inst_stringByPaddingToLength_withString_startingAtIndex_(
		unsafe.Pointer(x.Pointer()),
		C.ulong(newLength),
		objc.RefPointer(padString),
		C.ulong(padIndex),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) ComponentsSeparatedByString_(
	separator NSStringRef,
) (
	r0 NSArray,
) {
	ret := C.NSString_inst_componentsSeparatedByString_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(separator),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSString) SubstringFromIndex_(
	from NSUInteger,
) (
	r0 NSString,
) {
	ret := C.NSString_inst_substringFromIndex_(
		unsafe.Pointer(x.Pointer()),
		C.ulong(from),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) SubstringToIndex_(
	to NSUInteger,
) (
	r0 NSString,
) {
	ret := C.NSString_inst_substringToIndex_(
		unsafe.Pointer(x.Pointer()),
		C.ulong(to),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) ContainsString_(
	str NSStringRef,
) (
	r0 bool,
) {
	ret := C.NSString_inst_containsString_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(str),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSString) LocalizedCaseInsensitiveContainsString_(
	str NSStringRef,
) (
	r0 bool,
) {
	ret := C.NSString_inst_localizedCaseInsensitiveContainsString_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(str),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSString) LocalizedStandardContainsString_(
	str NSStringRef,
) (
	r0 bool,
) {
	ret := C.NSString_inst_localizedStandardContainsString_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(str),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSString) StringByReplacingOccurrencesOfString_withString_(
	target NSStringRef,
	replacement NSStringRef,
) (
	r0 NSString,
) {
	ret := C.NSString_inst_stringByReplacingOccurrencesOfString_withString_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(target),
		objc.RefPointer(replacement),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) WriteToFile_atomically_encoding_error_(
	path NSStringRef,
	useAuxiliaryFile bool,
	enc NSStringEncoding,
	error NSErrorRef,
) (
	r0 bool,
) {
	ret := C.NSString_inst_writeToFile_atomically_encoding_error_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(path),
		convertToObjCBool(useAuxiliaryFile),
		C.ulong(enc),
		objc.RefPointer(error),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSString) WriteToURL_atomically_encoding_error_(
	url NSURLRef,
	useAuxiliaryFile bool,
	enc NSStringEncoding,
	error NSErrorRef,
) (
	r0 bool,
) {
	ret := C.NSString_inst_writeToURL_atomically_encoding_error_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(url),
		convertToObjCBool(useAuxiliaryFile),
		C.ulong(enc),
		objc.RefPointer(error),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSString) PropertyList() (
	r0 objc.Object,
) {
	ret := C.NSString_inst_propertyList(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = objc.Object_fromPointer(ret)
	return
}

func (x gen_NSString) PropertyListFromStringsFileFormat() (
	r0 NSDictionary,
) {
	ret := C.NSString_inst_propertyListFromStringsFileFormat(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSDictionary_fromPointer(ret)
	return
}

func (x gen_NSString) DrawInRect_withAttributes_(
	rect NSRect,
	attrs NSDictionaryRef,
) {
	C.NSString_inst_drawInRect_withAttributes_(
		unsafe.Pointer(x.Pointer()),
		*(*C.NSRect)(unsafe.Pointer(&rect)),
		objc.RefPointer(attrs),
	)
	return
}

func (x gen_NSString) SizeWithAttributes_(
	attrs NSDictionaryRef,
) (
	r0 NSSize,
) {
	ret := C.NSString_inst_sizeWithAttributes_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(attrs),
	)
	r0 = *(*NSSize)(unsafe.Pointer(&ret))
	return
}

func (x gen_NSString) VariantFittingPresentationWidth_(
	width NSInteger,
) (
	r0 NSString,
) {
	ret := C.NSString_inst_variantFittingPresentationWidth_(
		unsafe.Pointer(x.Pointer()),
		C.long(width),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) CanBeConvertedToEncoding_(
	encoding NSStringEncoding,
) (
	r0 bool,
) {
	ret := C.NSString_inst_canBeConvertedToEncoding_(
		unsafe.Pointer(x.Pointer()),
		C.ulong(encoding),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSString) DataUsingEncoding_(
	encoding NSStringEncoding,
) (
	r0 NSData,
) {
	ret := C.NSString_inst_dataUsingEncoding_(
		unsafe.Pointer(x.Pointer()),
		C.ulong(encoding),
	)
	r0 = NSData_fromPointer(ret)
	return
}

func (x gen_NSString) DataUsingEncoding_allowLossyConversion_(
	encoding NSStringEncoding,
	lossy bool,
) (
	r0 NSData,
) {
	ret := C.NSString_inst_dataUsingEncoding_allowLossyConversion_(
		unsafe.Pointer(x.Pointer()),
		C.ulong(encoding),
		convertToObjCBool(lossy),
	)
	r0 = NSData_fromPointer(ret)
	return
}

func (x gen_NSString) CompletePathIntoString_caseSensitive_matchesIntoArray_filterTypes_(
	outputName NSStringRef,
	flag bool,
	outputArray NSArrayRef,
	filterTypes NSArrayRef,
) (
	r0 NSUInteger,
) {
	ret := C.NSString_inst_completePathIntoString_caseSensitive_matchesIntoArray_filterTypes_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(outputName),
		convertToObjCBool(flag),
		objc.RefPointer(outputArray),
		objc.RefPointer(filterTypes),
	)
	r0 = NSUInteger(ret)
	return
}

func (x gen_NSString) StringByAppendingPathComponent_(
	str NSStringRef,
) (
	r0 NSString,
) {
	ret := C.NSString_inst_stringByAppendingPathComponent_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(str),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) StringByAppendingPathExtension_(
	str NSStringRef,
) (
	r0 NSString,
) {
	ret := C.NSString_inst_stringByAppendingPathExtension_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(str),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) StringsByAppendingPaths_(
	paths NSArrayRef,
) (
	r0 NSArray,
) {
	ret := C.NSString_inst_stringsByAppendingPaths_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(paths),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSString) Length() (
	r0 NSUInteger,
) {
	ret := C.NSString_inst_length(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSUInteger(ret)
	return
}

func (x gen_NSString) Hash() (
	r0 NSUInteger,
) {
	ret := C.NSString_inst_hash(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSUInteger(ret)
	return
}

func (x gen_NSString) LowercaseString() (
	r0 NSString,
) {
	ret := C.NSString_inst_lowercaseString(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) LocalizedLowercaseString() (
	r0 NSString,
) {
	ret := C.NSString_inst_localizedLowercaseString(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) UppercaseString() (
	r0 NSString,
) {
	ret := C.NSString_inst_uppercaseString(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) LocalizedUppercaseString() (
	r0 NSString,
) {
	ret := C.NSString_inst_localizedUppercaseString(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) CapitalizedString() (
	r0 NSString,
) {
	ret := C.NSString_inst_capitalizedString(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) LocalizedCapitalizedString() (
	r0 NSString,
) {
	ret := C.NSString_inst_localizedCapitalizedString(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) DecomposedStringWithCanonicalMapping() (
	r0 NSString,
) {
	ret := C.NSString_inst_decomposedStringWithCanonicalMapping(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) DecomposedStringWithCompatibilityMapping() (
	r0 NSString,
) {
	ret := C.NSString_inst_decomposedStringWithCompatibilityMapping(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) PrecomposedStringWithCanonicalMapping() (
	r0 NSString,
) {
	ret := C.NSString_inst_precomposedStringWithCanonicalMapping(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) PrecomposedStringWithCompatibilityMapping() (
	r0 NSString,
) {
	ret := C.NSString_inst_precomposedStringWithCompatibilityMapping(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) DoubleValue() (
	r0 float64,
) {
	ret := C.NSString_inst_doubleValue(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float64(ret)
	return
}

func (x gen_NSString) FloatValue() (
	r0 float32,
) {
	ret := C.NSString_inst_floatValue(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float32(ret)
	return
}

func (x gen_NSString) IntValue() (
	r0 int32,
) {
	ret := C.NSString_inst_intValue(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = int32(ret)
	return
}

func (x gen_NSString) IntegerValue() (
	r0 NSInteger,
) {
	ret := C.NSString_inst_integerValue(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSInteger(ret)
	return
}

func (x gen_NSString) BoolValue() (
	r0 bool,
) {
	ret := C.NSString_inst_boolValue(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSString) Description() (
	r0 NSString,
) {
	ret := C.NSString_inst_description(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) FastestEncoding() (
	r0 NSStringEncoding,
) {
	ret := C.NSString_inst_fastestEncoding(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSStringEncoding(ret)
	return
}

func (x gen_NSString) SmallestEncoding() (
	r0 NSStringEncoding,
) {
	ret := C.NSString_inst_smallestEncoding(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSStringEncoding(ret)
	return
}

func (x gen_NSString) PathComponents() (
	r0 NSArray,
) {
	ret := C.NSString_inst_pathComponents(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSString) IsAbsolutePath() (
	r0 bool,
) {
	ret := C.NSString_inst_isAbsolutePath(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSString) LastPathComponent() (
	r0 NSString,
) {
	ret := C.NSString_inst_lastPathComponent(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) PathExtension() (
	r0 NSString,
) {
	ret := C.NSString_inst_pathExtension(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) StringByAbbreviatingWithTildeInPath() (
	r0 NSString,
) {
	ret := C.NSString_inst_stringByAbbreviatingWithTildeInPath(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) StringByDeletingLastPathComponent() (
	r0 NSString,
) {
	ret := C.NSString_inst_stringByDeletingLastPathComponent(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) StringByDeletingPathExtension() (
	r0 NSString,
) {
	ret := C.NSString_inst_stringByDeletingPathExtension(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) StringByExpandingTildeInPath() (
	r0 NSString,
) {
	ret := C.NSString_inst_stringByExpandingTildeInPath(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) StringByResolvingSymlinksInPath() (
	r0 NSString,
) {
	ret := C.NSString_inst_stringByResolvingSymlinksInPath(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) StringByStandardizingPath() (
	r0 NSString,
) {
	ret := C.NSString_inst_stringByStandardizingPath(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSString) StringByRemovingPercentEncoding() (
	r0 NSString,
) {
	ret := C.NSString_inst_stringByRemovingPercentEncoding(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

type NSErrorRef interface {
	Pointer() uintptr
	Init_asNSError() NSError
}

type gen_NSError struct {
	objc.Object
}

func NSError_fromPointer(ptr unsafe.Pointer) NSError {
	return NSError{gen_NSError{
		objc.Object_fromPointer(ptr),
	}}
}

func NSError_fromRef(ref objc.Ref) NSError {
	return NSError_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_NSError) InitWithDomain_code_userInfo__asNSError(
	domain NSStringRef,
	code NSInteger,
	dict NSDictionaryRef,
) (
	r0 NSError,
) {
	ret := C.NSError_inst_initWithDomain_code_userInfo_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(domain),
		C.long(code),
		objc.RefPointer(dict),
	)
	r0 = NSError_fromPointer(ret)
	return
}

func (x gen_NSError) Init_asNSError() (
	r0 NSError,
) {
	ret := C.NSError_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSError_fromPointer(ret)
	return
}

func (x gen_NSError) Code() (
	r0 NSInteger,
) {
	ret := C.NSError_inst_code(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSInteger(ret)
	return
}

func (x gen_NSError) Domain() (
	r0 NSString,
) {
	ret := C.NSError_inst_domain(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSError) UserInfo() (
	r0 NSDictionary,
) {
	ret := C.NSError_inst_userInfo(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSDictionary_fromPointer(ret)
	return
}

func (x gen_NSError) LocalizedDescription() (
	r0 NSString,
) {
	ret := C.NSError_inst_localizedDescription(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSError) LocalizedRecoveryOptions() (
	r0 NSArray,
) {
	ret := C.NSError_inst_localizedRecoveryOptions(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSError) LocalizedRecoverySuggestion() (
	r0 NSString,
) {
	ret := C.NSError_inst_localizedRecoverySuggestion(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSError) LocalizedFailureReason() (
	r0 NSString,
) {
	ret := C.NSError_inst_localizedFailureReason(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSError) RecoveryAttempter() (
	r0 objc.Object,
) {
	ret := C.NSError_inst_recoveryAttempter(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = objc.Object_fromPointer(ret)
	return
}

func (x gen_NSError) HelpAnchor() (
	r0 NSString,
) {
	ret := C.NSError_inst_helpAnchor(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSError) UnderlyingErrors() (
	r0 NSArray,
) {
	ret := C.NSError_inst_underlyingErrors(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

type NSThreadRef interface {
	Pointer() uintptr
	Init_asNSThread() NSThread
}

type gen_NSThread struct {
	objc.Object
}

func NSThread_fromPointer(ptr unsafe.Pointer) NSThread {
	return NSThread{gen_NSThread{
		objc.Object_fromPointer(ptr),
	}}
}

func NSThread_fromRef(ref objc.Ref) NSThread {
	return NSThread_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_NSThread) Init_asNSThread() (
	r0 NSThread,
) {
	ret := C.NSThread_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSThread_fromPointer(ret)
	return
}

func (x gen_NSThread) InitWithTarget_selector_object__asNSThread(
	target objc.Ref,
	selector objc.Selector,
	argument objc.Ref,
) (
	r0 NSThread,
) {
	ret := C.NSThread_inst_initWithTarget_selector_object_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(target),
		selector.SelectorAddress(),
		objc.RefPointer(argument),
	)
	r0 = NSThread_fromPointer(ret)
	return
}

func (x gen_NSThread) Start() {
	C.NSThread_inst_start(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_NSThread) Main() {
	C.NSThread_inst_main(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_NSThread) Cancel() {
	C.NSThread_inst_cancel(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_NSThread) IsExecuting() (
	r0 bool,
) {
	ret := C.NSThread_inst_isExecuting(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSThread) IsFinished() (
	r0 bool,
) {
	ret := C.NSThread_inst_isFinished(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSThread) IsCancelled() (
	r0 bool,
) {
	ret := C.NSThread_inst_isCancelled(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSThread) IsMainThread() (
	r0 bool,
) {
	ret := C.NSThread_inst_isMainThread(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSThread) Name() (
	r0 NSString,
) {
	ret := C.NSThread_inst_name(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSThread) SetName_(
	value NSStringRef,
) {
	C.NSThread_inst_setName_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

func (x gen_NSThread) StackSize() (
	r0 NSUInteger,
) {
	ret := C.NSThread_inst_stackSize(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSUInteger(ret)
	return
}

func (x gen_NSThread) SetStackSize_(
	value NSUInteger,
) {
	C.NSThread_inst_setStackSize_(
		unsafe.Pointer(x.Pointer()),
		C.ulong(value),
	)
	return
}

func (x gen_NSThread) ThreadPriority() (
	r0 float64,
) {
	ret := C.NSThread_inst_threadPriority(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float64(ret)
	return
}

func (x gen_NSThread) SetThreadPriority_(
	value float64,
) {
	C.NSThread_inst_setThreadPriority_(
		unsafe.Pointer(x.Pointer()),
		C.double(value),
	)
	return
}

type NSURLRef interface {
	Pointer() uintptr
	Init_asNSURL() NSURL
}

type gen_NSURL struct {
	objc.Object
}

func NSURL_fromPointer(ptr unsafe.Pointer) NSURL {
	return NSURL{gen_NSURL{
		objc.Object_fromPointer(ptr),
	}}
}

func NSURL_fromRef(ref objc.Ref) NSURL {
	return NSURL_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_NSURL) InitWithString__asNSURL(
	URLString NSStringRef,
) (
	r0 NSURL,
) {
	ret := C.NSURL_inst_initWithString_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(URLString),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) InitWithString_relativeToURL__asNSURL(
	URLString NSStringRef,
	baseURL NSURLRef,
) (
	r0 NSURL,
) {
	ret := C.NSURL_inst_initWithString_relativeToURL_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(URLString),
		objc.RefPointer(baseURL),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) InitFileURLWithPath_isDirectory__asNSURL(
	path NSStringRef,
	isDir bool,
) (
	r0 NSURL,
) {
	ret := C.NSURL_inst_initFileURLWithPath_isDirectory_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(path),
		convertToObjCBool(isDir),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) InitFileURLWithPath_relativeToURL__asNSURL(
	path NSStringRef,
	baseURL NSURLRef,
) (
	r0 NSURL,
) {
	ret := C.NSURL_inst_initFileURLWithPath_relativeToURL_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(path),
		objc.RefPointer(baseURL),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) InitFileURLWithPath_isDirectory_relativeToURL__asNSURL(
	path NSStringRef,
	isDir bool,
	baseURL NSURLRef,
) (
	r0 NSURL,
) {
	ret := C.NSURL_inst_initFileURLWithPath_isDirectory_relativeToURL_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(path),
		convertToObjCBool(isDir),
		objc.RefPointer(baseURL),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) InitFileURLWithPath__asNSURL(
	path NSStringRef,
) (
	r0 NSURL,
) {
	ret := C.NSURL_inst_initFileURLWithPath_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(path),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) InitAbsoluteURLWithDataRepresentation_relativeToURL__asNSURL(
	data NSDataRef,
	baseURL NSURLRef,
) (
	r0 NSURL,
) {
	ret := C.NSURL_inst_initAbsoluteURLWithDataRepresentation_relativeToURL_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(data),
		objc.RefPointer(baseURL),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) InitWithDataRepresentation_relativeToURL__asNSURL(
	data NSDataRef,
	baseURL NSURLRef,
) (
	r0 NSURL,
) {
	ret := C.NSURL_inst_initWithDataRepresentation_relativeToURL_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(data),
		objc.RefPointer(baseURL),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) IsEqual_(
	anObject objc.Ref,
) (
	r0 bool,
) {
	ret := C.NSURL_inst_isEqual_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(anObject),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSURL) CheckResourceIsReachableAndReturnError_(
	error NSErrorRef,
) (
	r0 bool,
) {
	ret := C.NSURL_inst_checkResourceIsReachableAndReturnError_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(error),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSURL) IsFileReferenceURL() (
	r0 bool,
) {
	ret := C.NSURL_inst_isFileReferenceURL(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSURL) ResourceValuesForKeys_error_(
	keys NSArrayRef,
	error NSErrorRef,
) (
	r0 NSDictionary,
) {
	ret := C.NSURL_inst_resourceValuesForKeys_error_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(keys),
		objc.RefPointer(error),
	)
	r0 = NSDictionary_fromPointer(ret)
	return
}

func (x gen_NSURL) SetResourceValues_error_(
	keyedValues NSDictionaryRef,
	error NSErrorRef,
) (
	r0 bool,
) {
	ret := C.NSURL_inst_setResourceValues_error_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(keyedValues),
		objc.RefPointer(error),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSURL) RemoveAllCachedResourceValues() {
	C.NSURL_inst_removeAllCachedResourceValues(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_NSURL) FileReferenceURL() (
	r0 NSURL,
) {
	ret := C.NSURL_inst_fileReferenceURL(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) URLByAppendingPathComponent_(
	pathComponent NSStringRef,
) (
	r0 NSURL,
) {
	ret := C.NSURL_inst_URLByAppendingPathComponent_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(pathComponent),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) URLByAppendingPathComponent_isDirectory_(
	pathComponent NSStringRef,
	isDirectory bool,
) (
	r0 NSURL,
) {
	ret := C.NSURL_inst_URLByAppendingPathComponent_isDirectory_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(pathComponent),
		convertToObjCBool(isDirectory),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) URLByAppendingPathExtension_(
	pathExtension NSStringRef,
) (
	r0 NSURL,
) {
	ret := C.NSURL_inst_URLByAppendingPathExtension_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(pathExtension),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) StartAccessingSecurityScopedResource() (
	r0 bool,
) {
	ret := C.NSURL_inst_startAccessingSecurityScopedResource(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSURL) StopAccessingSecurityScopedResource() {
	C.NSURL_inst_stopAccessingSecurityScopedResource(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_NSURL) CheckPromisedItemIsReachableAndReturnError_(
	error NSErrorRef,
) (
	r0 bool,
) {
	ret := C.NSURL_inst_checkPromisedItemIsReachableAndReturnError_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(error),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSURL) PromisedItemResourceValuesForKeys_error_(
	keys NSArrayRef,
	error NSErrorRef,
) (
	r0 NSDictionary,
) {
	ret := C.NSURL_inst_promisedItemResourceValuesForKeys_error_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(keys),
		objc.RefPointer(error),
	)
	r0 = NSDictionary_fromPointer(ret)
	return
}

func (x gen_NSURL) Init_asNSURL() (
	r0 NSURL,
) {
	ret := C.NSURL_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) DataRepresentation() (
	r0 NSData,
) {
	ret := C.NSURL_inst_dataRepresentation(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSData_fromPointer(ret)
	return
}

func (x gen_NSURL) IsFileURL() (
	r0 bool,
) {
	ret := C.NSURL_inst_isFileURL(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSURL) AbsoluteString() (
	r0 NSString,
) {
	ret := C.NSURL_inst_absoluteString(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSURL) AbsoluteURL() (
	r0 NSURL,
) {
	ret := C.NSURL_inst_absoluteURL(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) BaseURL() (
	r0 NSURL,
) {
	ret := C.NSURL_inst_baseURL(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) Fragment() (
	r0 NSString,
) {
	ret := C.NSURL_inst_fragment(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSURL) Host() (
	r0 NSString,
) {
	ret := C.NSURL_inst_host(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSURL) LastPathComponent() (
	r0 NSString,
) {
	ret := C.NSURL_inst_lastPathComponent(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSURL) Password() (
	r0 NSString,
) {
	ret := C.NSURL_inst_password(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSURL) Path() (
	r0 NSString,
) {
	ret := C.NSURL_inst_path(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSURL) PathComponents() (
	r0 NSArray,
) {
	ret := C.NSURL_inst_pathComponents(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSArray_fromPointer(ret)
	return
}

func (x gen_NSURL) PathExtension() (
	r0 NSString,
) {
	ret := C.NSURL_inst_pathExtension(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSURL) Port() (
	r0 NSNumber,
) {
	ret := C.NSURL_inst_port(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSNumber_fromPointer(ret)
	return
}

func (x gen_NSURL) Query() (
	r0 NSString,
) {
	ret := C.NSURL_inst_query(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSURL) RelativePath() (
	r0 NSString,
) {
	ret := C.NSURL_inst_relativePath(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSURL) RelativeString() (
	r0 NSString,
) {
	ret := C.NSURL_inst_relativeString(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSURL) ResourceSpecifier() (
	r0 NSString,
) {
	ret := C.NSURL_inst_resourceSpecifier(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSURL) Scheme() (
	r0 NSString,
) {
	ret := C.NSURL_inst_scheme(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSURL) StandardizedURL() (
	r0 NSURL,
) {
	ret := C.NSURL_inst_standardizedURL(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) User() (
	r0 NSString,
) {
	ret := C.NSURL_inst_user(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSURL) FilePathURL() (
	r0 NSURL,
) {
	ret := C.NSURL_inst_filePathURL(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) URLByDeletingLastPathComponent() (
	r0 NSURL,
) {
	ret := C.NSURL_inst_URLByDeletingLastPathComponent(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) URLByDeletingPathExtension() (
	r0 NSURL,
) {
	ret := C.NSURL_inst_URLByDeletingPathExtension(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) URLByResolvingSymlinksInPath() (
	r0 NSURL,
) {
	ret := C.NSURL_inst_URLByResolvingSymlinksInPath(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) URLByStandardizingPath() (
	r0 NSURL,
) {
	ret := C.NSURL_inst_URLByStandardizingPath(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURL) HasDirectoryPath() (
	r0 bool,
) {
	ret := C.NSURL_inst_hasDirectoryPath(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

type NSURLRequestRef interface {
	Pointer() uintptr
	Init_asNSURLRequest() NSURLRequest
}

type gen_NSURLRequest struct {
	objc.Object
}

func NSURLRequest_fromPointer(ptr unsafe.Pointer) NSURLRequest {
	return NSURLRequest{gen_NSURLRequest{
		objc.Object_fromPointer(ptr),
	}}
}

func NSURLRequest_fromRef(ref objc.Ref) NSURLRequest {
	return NSURLRequest_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_NSURLRequest) InitWithURL__asNSURLRequest(
	URL NSURLRef,
) (
	r0 NSURLRequest,
) {
	ret := C.NSURLRequest_inst_initWithURL_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(URL),
	)
	r0 = NSURLRequest_fromPointer(ret)
	return
}

func (x gen_NSURLRequest) ValueForHTTPHeaderField_(
	field NSStringRef,
) (
	r0 NSString,
) {
	ret := C.NSURLRequest_inst_valueForHTTPHeaderField_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(field),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSURLRequest) Init_asNSURLRequest() (
	r0 NSURLRequest,
) {
	ret := C.NSURLRequest_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSURLRequest_fromPointer(ret)
	return
}

func (x gen_NSURLRequest) HTTPMethod() (
	r0 NSString,
) {
	ret := C.NSURLRequest_inst_HTTPMethod(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSURLRequest) URL() (
	r0 NSURL,
) {
	ret := C.NSURLRequest_inst_URL(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURLRequest) HTTPBody() (
	r0 NSData,
) {
	ret := C.NSURLRequest_inst_HTTPBody(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSData_fromPointer(ret)
	return
}

func (x gen_NSURLRequest) MainDocumentURL() (
	r0 NSURL,
) {
	ret := C.NSURLRequest_inst_mainDocumentURL(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSURL_fromPointer(ret)
	return
}

func (x gen_NSURLRequest) AllHTTPHeaderFields() (
	r0 NSDictionary,
) {
	ret := C.NSURLRequest_inst_allHTTPHeaderFields(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSDictionary_fromPointer(ret)
	return
}

func (x gen_NSURLRequest) TimeoutInterval() (
	r0 float64,
) {
	ret := C.NSURLRequest_inst_timeoutInterval(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = float64(ret)
	return
}

func (x gen_NSURLRequest) HTTPShouldHandleCookies() (
	r0 bool,
) {
	ret := C.NSURLRequest_inst_HTTPShouldHandleCookies(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSURLRequest) HTTPShouldUsePipelining() (
	r0 bool,
) {
	ret := C.NSURLRequest_inst_HTTPShouldUsePipelining(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSURLRequest) AllowsCellularAccess() (
	r0 bool,
) {
	ret := C.NSURLRequest_inst_allowsCellularAccess(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSURLRequest) AllowsConstrainedNetworkAccess() (
	r0 bool,
) {
	ret := C.NSURLRequest_inst_allowsConstrainedNetworkAccess(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSURLRequest) AllowsExpensiveNetworkAccess() (
	r0 bool,
) {
	ret := C.NSURLRequest_inst_allowsExpensiveNetworkAccess(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSURLRequest) AssumesHTTP3Capable() (
	r0 bool,
) {
	ret := C.NSURLRequest_inst_assumesHTTP3Capable(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

type NSNotificationRef interface {
	Pointer() uintptr
	Init_asNSNotification() NSNotification
}

type gen_NSNotification struct {
	objc.Object
}

func NSNotification_fromPointer(ptr unsafe.Pointer) NSNotification {
	return NSNotification{gen_NSNotification{
		objc.Object_fromPointer(ptr),
	}}
}

func NSNotification_fromRef(ref objc.Ref) NSNotification {
	return NSNotification_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_NSNotification) Init_asNSNotification() (
	r0 NSNotification,
) {
	ret := C.NSNotification_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSNotification_fromPointer(ret)
	return
}

func (x gen_NSNotification) Object_() (
	r0 objc.Object,
) {
	ret := C.NSNotification_inst_object_(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = objc.Object_fromPointer(ret)
	return
}

func (x gen_NSNotification) UserInfo() (
	r0 NSDictionary,
) {
	ret := C.NSNotification_inst_userInfo(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSDictionary_fromPointer(ret)
	return
}

type NSOperationQueueRef interface {
	Pointer() uintptr
	Init_asNSOperationQueue() NSOperationQueue
}

type gen_NSOperationQueue struct {
	objc.Object
}

func NSOperationQueue_fromPointer(ptr unsafe.Pointer) NSOperationQueue {
	return NSOperationQueue{gen_NSOperationQueue{
		objc.Object_fromPointer(ptr),
	}}
}

func NSOperationQueue_fromRef(ref objc.Ref) NSOperationQueue {
	return NSOperationQueue_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_NSOperationQueue) AddOperations_waitUntilFinished_(
	ops NSArrayRef,
	wait bool,
) {
	C.NSOperationQueue_inst_addOperations_waitUntilFinished_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(ops),
		convertToObjCBool(wait),
	)
	return
}

func (x gen_NSOperationQueue) CancelAllOperations() {
	C.NSOperationQueue_inst_cancelAllOperations(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_NSOperationQueue) WaitUntilAllOperationsAreFinished() {
	C.NSOperationQueue_inst_waitUntilAllOperationsAreFinished(
		unsafe.Pointer(x.Pointer()),
	)
	return
}

func (x gen_NSOperationQueue) Init_asNSOperationQueue() (
	r0 NSOperationQueue,
) {
	ret := C.NSOperationQueue_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSOperationQueue_fromPointer(ret)
	return
}

func (x gen_NSOperationQueue) MaxConcurrentOperationCount() (
	r0 NSInteger,
) {
	ret := C.NSOperationQueue_inst_maxConcurrentOperationCount(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSInteger(ret)
	return
}

func (x gen_NSOperationQueue) SetMaxConcurrentOperationCount_(
	value NSInteger,
) {
	C.NSOperationQueue_inst_setMaxConcurrentOperationCount_(
		unsafe.Pointer(x.Pointer()),
		C.long(value),
	)
	return
}

func (x gen_NSOperationQueue) IsSuspended() (
	r0 bool,
) {
	ret := C.NSOperationQueue_inst_isSuspended(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = convertObjCBoolToGo(ret)
	return
}

func (x gen_NSOperationQueue) SetSuspended_(
	value bool,
) {
	C.NSOperationQueue_inst_setSuspended_(
		unsafe.Pointer(x.Pointer()),
		convertToObjCBool(value),
	)
	return
}

func (x gen_NSOperationQueue) Name() (
	r0 NSString,
) {
	ret := C.NSOperationQueue_inst_name(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSString_fromPointer(ret)
	return
}

func (x gen_NSOperationQueue) SetName_(
	value NSStringRef,
) {
	C.NSOperationQueue_inst_setName_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(value),
	)
	return
}

type NSNotificationCenterRef interface {
	Pointer() uintptr
	Init_asNSNotificationCenter() NSNotificationCenter
}

type gen_NSNotificationCenter struct {
	objc.Object
}

func NSNotificationCenter_fromPointer(ptr unsafe.Pointer) NSNotificationCenter {
	return NSNotificationCenter{gen_NSNotificationCenter{
		objc.Object_fromPointer(ptr),
	}}
}

func NSNotificationCenter_fromRef(ref objc.Ref) NSNotificationCenter {
	return NSNotificationCenter_fromPointer(unsafe.Pointer(ref.Pointer()))
}

func (x gen_NSNotificationCenter) AddObserver_selector_name_object_(
	observer objc.Ref,
	aSelector objc.Selector,
	aName NSStringRef,
	anObject objc.Ref,
) {
	C.NSNotificationCenter_inst_addObserver_selector_name_object_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(observer),
		aSelector.SelectorAddress(),
		objc.RefPointer(aName),
		objc.RefPointer(anObject),
	)
	return
}

func (x gen_NSNotificationCenter) RemoveObserver_name_object_(
	observer objc.Ref,
	aName NSStringRef,
	anObject objc.Ref,
) {
	C.NSNotificationCenter_inst_removeObserver_name_object_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(observer),
		objc.RefPointer(aName),
		objc.RefPointer(anObject),
	)
	return
}

func (x gen_NSNotificationCenter) RemoveObserver_(
	observer objc.Ref,
) {
	C.NSNotificationCenter_inst_removeObserver_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(observer),
	)
	return
}

func (x gen_NSNotificationCenter) PostNotification_(
	notification NSNotificationRef,
) {
	C.NSNotificationCenter_inst_postNotification_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(notification),
	)
	return
}

func (x gen_NSNotificationCenter) PostNotificationName_object_userInfo_(
	aName NSStringRef,
	anObject objc.Ref,
	aUserInfo NSDictionaryRef,
) {
	C.NSNotificationCenter_inst_postNotificationName_object_userInfo_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(aName),
		objc.RefPointer(anObject),
		objc.RefPointer(aUserInfo),
	)
	return
}

func (x gen_NSNotificationCenter) PostNotificationName_object_(
	aName NSStringRef,
	anObject objc.Ref,
) {
	C.NSNotificationCenter_inst_postNotificationName_object_(
		unsafe.Pointer(x.Pointer()),
		objc.RefPointer(aName),
		objc.RefPointer(anObject),
	)
	return
}

func (x gen_NSNotificationCenter) Init_asNSNotificationCenter() (
	r0 NSNotificationCenter,
) {
	ret := C.NSNotificationCenter_inst_init(
		unsafe.Pointer(x.Pointer()),
	)
	r0 = NSNotificationCenter_fromPointer(ret)
	return
}
