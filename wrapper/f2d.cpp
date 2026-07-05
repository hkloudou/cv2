// Feature-based localization (github.com/hkloudou/cv2/f2d).
//
// Compiled into libcv2f2dwrapper.a, shipped in the optional libs/<goos>_<goarch>_f2d
// modules; only builds that import the f2d subpackage link it (plus
// opencv_features2d and opencv_calib3d).

#include "cv2internal.h"

#include <opencv2/core/version.hpp>
#include <opencv2/imgproc.hpp>

// OpenCV 5 renamed features2d -> features and moved findHomography/RANSAC
// into the new geometry module. The 5.x compat headers (features2d.hpp,
// calib3d.hpp) forward to modules we deliberately do not build (stereo,
// calib), so include the precise per-generation headers instead.
#if CV_VERSION_MAJOR >= 5
#include <opencv2/features.hpp>
#include <opencv2/geometry.hpp>
#else
#include <opencv2/calib3d.hpp>
#include <opencv2/features2d.hpp>
#endif

#include <vector>

extern "C" char *Cv2_FeatureLocate(Cv2Mat parent, Cv2Mat sub, int maxFeatures,
                                   double ratio, double ransacThreshold, int minInliers,
                                   double *corners /* 8 doubles: TL TR BR BL */,
                                   int *inliers, int *matched, int *found);

char *Cv2_FeatureLocate(Cv2Mat parent, Cv2Mat sub, int maxFeatures,
                        double ratio, double ransacThreshold, int minInliers,
                        double *corners, int *inliers, int *matched, int *found)
{
  if (parent == nullptr || sub == nullptr || corners == nullptr ||
      inliers == nullptr || matched == nullptr || found == nullptr)
  {
    return Cv2_CopyError("null argument");
  }
  *found = 0;
  *inliers = 0;
  *matched = 0;
  try
  {
    // Work in grayscale; inputs from ImageToMatRGBA are CV_8UC4 (RGBA).
    cv::Mat parentGray, subGray;
    if (parent->channels() == 4)
      cv::cvtColor(*parent, parentGray, cv::COLOR_RGBA2GRAY);
    else if (parent->channels() == 3)
      cv::cvtColor(*parent, parentGray, cv::COLOR_RGB2GRAY);
    else
      parentGray = *parent;
    if (sub->channels() == 4)
      cv::cvtColor(*sub, subGray, cv::COLOR_RGBA2GRAY);
    else if (sub->channels() == 3)
      cv::cvtColor(*sub, subGray, cv::COLOR_RGB2GRAY);
    else
      subGray = *sub;

    cv::Ptr<cv::ORB> orb = cv::ORB::create(maxFeatures);
    std::vector<cv::KeyPoint> parentKp, subKp;
    cv::Mat parentDesc, subDesc;
    orb->detectAndCompute(parentGray, cv::noArray(), parentKp, parentDesc);
    orb->detectAndCompute(subGray, cv::noArray(), subKp, subDesc);
    if (parentDesc.empty() || subDesc.empty() || subKp.size() < 4)
    {
      return nullptr; // not found: not enough structure to match
    }

    cv::BFMatcher matcher(cv::NORM_HAMMING);
    std::vector<std::vector<cv::DMatch>> knn;
    matcher.knnMatch(subDesc, parentDesc, knn, 2);

    std::vector<cv::Point2f> subPts, parentPts;
    for (size_t i = 0; i < knn.size(); i++)
    {
      if (knn[i].size() == 2 && knn[i][0].distance < ratio * knn[i][1].distance)
      {
        subPts.push_back(subKp[knn[i][0].queryIdx].pt);
        parentPts.push_back(parentKp[knn[i][0].trainIdx].pt);
      }
    }
    *matched = (int)subPts.size();
    if ((int)subPts.size() < 4 || (int)subPts.size() < minInliers)
    {
      return nullptr; // not found
    }

    std::vector<unsigned char> inlierMask;
    cv::Mat H = cv::findHomography(subPts, parentPts, cv::RANSAC, ransacThreshold, inlierMask);
    if (H.empty())
    {
      return nullptr; // not found
    }
    int inlierCount = 0;
    for (size_t i = 0; i < inlierMask.size(); i++)
    {
      if (inlierMask[i])
        inlierCount++;
    }
    *inliers = inlierCount;
    if (inlierCount < minInliers)
    {
      return nullptr; // not found: consensus too weak
    }

    const float w = (float)subGray.cols, h = (float)subGray.rows;
    std::vector<cv::Point2f> quad(4), projected;
    quad[0] = cv::Point2f(0, 0);
    quad[1] = cv::Point2f(w, 0);
    quad[2] = cv::Point2f(w, h);
    quad[3] = cv::Point2f(0, h);
    cv::perspectiveTransform(quad, projected, H);
    for (int i = 0; i < 4; i++)
    {
      corners[i * 2] = projected[i].x;
      corners[i * 2 + 1] = projected[i].y;
    }
    *found = 1;
    return nullptr;
  }
  catch (...)
  {
    return cv2_current_exception_message();
  }
}
