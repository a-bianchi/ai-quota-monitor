// Screen content: one page per scope (or more when it does not fit), rotating.
#pragma once

void uiBegin();
// Call from loop(): picks up new data and rotates the pages.
void uiLoop();
