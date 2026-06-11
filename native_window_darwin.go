//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework QuartzCore
#import <Cocoa/Cocoa.h>
#import <QuartzCore/QuartzCore.h>

static void applyNativeWindowStyleDarwin(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSWindow *window = [NSApp mainWindow];
        if (window == nil) {
            window = [NSApp keyWindow];
        }
        if (window == nil) {
            return;
        }

        window.opaque = NO;
        window.backgroundColor = [NSColor clearColor];
        window.hasShadow = YES;

        NSView *contentView = window.contentView;
        contentView.wantsLayer = YES;
        contentView.layer.backgroundColor = [NSColor clearColor].CGColor;

        for (NSView *subview in contentView.subviews) {
            if ([subview isKindOfClass:[NSVisualEffectView class]]) {
                subview.wantsLayer = YES;
                subview.layer.cornerRadius = 24.0;
                subview.layer.masksToBounds = YES;
                subview.layer.backgroundColor = [NSColor clearColor].CGColor;
            }
        }
    });
}
*/
import "C"

func applyNativeWindowStyle() {
	C.applyNativeWindowStyleDarwin()
}
