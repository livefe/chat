package com.nearby.chat.feature.post.ui

import androidx.compose.ui.graphics.Color
import com.nearby.chat.R
import com.nearby.chat.core.model.PostItem
import com.nearby.chat.core.post.PostListTag
import com.nearby.chat.core.post.PostListTagTone
import com.nearby.chat.core.post.PostSignupRules
import com.nearby.chat.core.post.PostStatus
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.ui.UiText

internal data class PostCornerBadge(
    val text: UiText,
    val bg: Color,
    val fg: Color,
) {
    companion object {
        fun owner(status: Int, phase: PostListTag?): PostCornerBadge? {
            if (status == PostStatus.PENDING || status == PostStatus.REJECTED || status == PostStatus.OFFLINE) {
                return ofStatus(status)
            }
            if (phase != null) return ofPhase(phase)
            if (status == PostStatus.PUBLISHED) return ofStatus(status)
            return null
        }

        fun joined(post: PostItem, phase: PostListTag?): PostCornerBadge? {
            when (post.status) {
                PostStatus.OFFLINE, PostStatus.PENDING, PostStatus.REJECTED -> return ofStatus(post.status)
            }
            when (post.mySignupStatus) {
                PostSignupRules.STATUS_WAITLISTED -> return PostCornerBadge(
                    UiText.of(R.string.signup_badge_waitlisted),
                    AppColors.WarningTint,
                    AppColors.WarningStrong,
                )

                PostSignupRules.STATUS_CONFIRMED -> checkInBadge(post)?.let { return it }
            }
            if (phase != null) return ofPhase(phase)
            if (post.mySignupStatus == PostSignupRules.STATUS_CONFIRMED) {
                return PostCornerBadge(UiText.of(R.string.signup_status_confirmed), AppColors.BrandTint, AppColors.Brand)
            }
            return null
        }

        private fun checkInBadge(post: PostItem): PostCornerBadge? {
            val needCheckIn = post.extras?.needSignup == true && post.extras.needCheckIn
            if (!needCheckIn) return null
            return when {
                PostSignupRules.hasTimestamp(post.myCheckedInAt) -> PostCornerBadge(
                    UiText.of(R.string.signup_checked_in),
                    AppColors.SuccessTint,
                    AppColors.SuccessStrong,
                )

                PostSignupRules.isEventEnded(post.endAt) -> PostCornerBadge(
                    UiText.of(R.string.signup_not_checked_in),
                    AppColors.OverlaySoft,
                    AppColors.TextTertiary,
                )

                PostSignupRules.isCheckInWindowOpen(post.startAt, post.endAt) -> PostCornerBadge(
                    UiText.of(R.string.signup_badge_check_in_pending),
                    AppColors.BrandTint,
                    AppColors.BrandDark,
                )

                else -> null
            }
        }

        private fun ofPhase(phase: PostListTag): PostCornerBadge = PostCornerBadge(phase.text, tagBg(phase.tone), tagFg(phase.tone))

        private fun ofStatus(status: Int): PostCornerBadge? {
            val label = when (status) {
                PostStatus.PENDING -> R.string.post_status_pending
                PostStatus.PUBLISHED -> R.string.post_status_published
                PostStatus.REJECTED -> R.string.post_status_rejected
                PostStatus.OFFLINE -> R.string.post_status_offline
                else -> return null
            }
            return PostCornerBadge(UiText.of(label), statusBg(status), statusFg(status))
        }

        private fun tagBg(tone: PostListTagTone): Color = when (tone) {
            PostListTagTone.PhasePending -> AppColors.BrandTint
            PostListTagTone.PhaseLive -> AppColors.SuccessTint
            PostListTagTone.PhaseDone -> AppColors.OverlaySoft
            PostListTagTone.Fee -> AppColors.WarningTintLight
            PostListTagTone.Signup -> AppColors.OverlaySoft
            PostListTagTone.None -> AppColors.OverlayHairline
        }

        private fun tagFg(tone: PostListTagTone): Color = when (tone) {
            PostListTagTone.PhasePending -> AppColors.BrandDark
            PostListTagTone.PhaseLive -> AppColors.SuccessStrong
            PostListTagTone.PhaseDone -> AppColors.TextTertiary
            PostListTagTone.Fee -> AppColors.WarningStrong
            PostListTagTone.Signup -> AppColors.BrandDark
            PostListTagTone.None -> AppColors.TextSecondary
        }

        private fun statusFg(status: Int): Color = when (status) {
            PostStatus.PENDING -> AppColors.WarningStrong
            PostStatus.PUBLISHED -> AppColors.Brand
            PostStatus.REJECTED -> AppColors.DangerStrong
            else -> AppColors.TextTertiary
        }

        private fun statusBg(status: Int): Color = when (status) {
            PostStatus.PENDING -> AppColors.WarningTint
            PostStatus.PUBLISHED -> AppColors.BrandTint
            PostStatus.REJECTED -> AppColors.DangerTint
            else -> AppColors.OverlaySoft
        }
    }
}
