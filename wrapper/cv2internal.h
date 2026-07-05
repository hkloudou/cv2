// Shared helpers for wrapper translation units beyond cv2capi.cpp.
// C++ only; not part of the C ABI.
#ifndef CV2_INTERNAL_H_
#define CV2_INTERNAL_H_

#include "cv2capi.h"

#include <exception>

// Converts the in-flight exception into an error string allocated through
// Cv2_CopyError (so Cv2_FreeString can release it). Call from a catch block.
inline char *cv2_current_exception_message()
{
  try
  {
    throw;
  }
  catch (const std::exception &e)
  {
    return Cv2_CopyError(e.what());
  }
  catch (...)
  {
    return Cv2_CopyError("unknown C++ exception");
  }
}

#endif // CV2_INTERNAL_H_
