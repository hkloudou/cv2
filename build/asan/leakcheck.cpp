// AddressSanitizer harness for the wrapper: exercises every base entry
// point, including error paths, in a loop; LeakSanitizer reports anything
// not released at exit. Built and run by the asan job in the test workflow:
//
//   g++ -fsanitize=address -I <sdk include> leakcheck.cpp \
//       -L <libs module dir> -lcv2wrapper -lopencv_imgproc -lopencv_core \
//       -lzlib -lstdc++ -lm -lpthread -ldl
#include "../../wrapper/cv2capi.h"

#include <cstdio>
#include <vector>

int main()
{
  const int w = 320, h = 240, tw = 48, th = 36;
  std::vector<char> parent(w * h * 4), templ(tw * th * 4), bad(100 * 10 * 4);
  for (size_t i = 0; i < parent.size(); i++)
    parent[i] = (char)(i * 2654435761u >> 13);
  for (size_t i = 0; i < templ.size(); i++)
    templ[i] = (char)(i * 40503u >> 7);
  for (size_t i = 0; i < bad.size(); i++)
    bad[i] = (char)i;

  for (int iter = 0; iter < 200; iter++)
  {
    Cv2ByteArray pb = {parent.data(), (int)parent.size()};
    Cv2ByteArray tb = {templ.data(), (int)templ.size()};
    Cv2Mat pm = Cv2_Mat_NewFromBytes(h, w, 24, pb); // CV_8UC4
    Cv2Mat tm = Cv2_Mat_NewFromBytes(th, tw, 24, tb);
    Cv2Mat result = Cv2_Mat_New();
    Cv2Mat mask = Cv2_Mat_New();
    char *err = Cv2_MatchTemplate(pm, tm, result, 5, mask);
    if (err)
    {
      std::printf("unexpected: %s\n", err);
      Cv2_FreeString(err);
      return 1;
    }
    double mn, mx;
    Cv2Point mnl, mxl;
    err = Cv2_MinMaxLoc(result, &mn, &mx, &mnl, &mxl);
    if (err)
    {
      std::printf("unexpected: %s\n", err);
      Cv2_FreeString(err);
      return 1;
    }

    // Data copy round trip.
    std::vector<char> out(parent.size());
    err = Cv2_Mat_DataCopy(pm, out.data(), (int)out.size());
    if (err)
    {
      std::printf("unexpected: %s\n", err);
      Cv2_FreeString(err);
      return 1;
    }

    // Error path: incompatible sizes (error string alloc/free cycle).
    Cv2ByteArray bb = {bad.data(), (int)bad.size()};
    Cv2Mat wide = Cv2_Mat_NewFromBytes(10, 100, 24, bb);
    Cv2Mat r2 = Cv2_Mat_New();
    err = Cv2_MatchTemplate(tm, wide, r2, 5, mask);
    if (err == nullptr)
    {
      std::printf("expected size error\n");
      return 1;
    }
    Cv2_FreeString(err);

    // Error path: null handles.
    err = Cv2_MatchTemplate(nullptr, tm, result, 5, mask);
    if (err == nullptr)
    {
      std::printf("expected null error\n");
      return 1;
    }
    Cv2_FreeString(err);

    // Error path: invalid constructor params -> NULL, nothing allocated.
    if (Cv2_Mat_NewFromBytes(-1, 10, 0, bb) != nullptr)
    {
      std::printf("expected NULL for invalid dims\n");
      return 1;
    }

    Cv2_Mat_Close(pm);
    Cv2_Mat_Close(tm);
    Cv2_Mat_Close(result);
    Cv2_Mat_Close(mask);
    Cv2_Mat_Close(wide);
    Cv2_Mat_Close(r2);
  }
  std::printf("leakcheck: 200 iterations clean\n");
  return 0;
}
