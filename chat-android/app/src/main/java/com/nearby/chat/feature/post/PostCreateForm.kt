package com.nearby.chat.feature.post

import androidx.annotation.StringRes
import com.nearby.chat.R
import com.nearby.chat.core.format.MoneyFormat
import com.nearby.chat.core.model.PostDraft
import com.nearby.chat.core.model.PostExtras
import com.nearby.chat.core.model.PostItem
import com.nearby.chat.core.post.PostCategories
import com.nearby.chat.core.post.PostSignupRules
import com.nearby.chat.core.ui.UiText
import java.util.UUID

internal object PostCreateForm {
    private const val FEE_TYPE_NONE = 0
    private const val FEE_TYPE_CHARGE = 1
    private const val FEE_TYPE_REWARD = 2

    fun validate(s: PostCreateUiState): UiText? {
        if (!s.locationReady && s.category.isNotBlank() && s.title.isNotBlank() &&
            (s.category == PostCategories.IDLE || s.content.isNotBlank())
        ) {
            return s.locationError ?: UiText.of(R.string.post_need_location)
        }
        return validateFields(s)?.let { UiText.of(it) }
    }

    @StringRes
    private fun validateFields(s: PostCreateUiState): Int? {
        if (s.category.isBlank()) return R.string.post_need_category
        if (s.title.isBlank()) return R.string.post_need_title
        if (s.category != PostCategories.IDLE && s.content.isBlank()) return R.string.post_need_content
        if (s.category == PostCategories.EVENT) validateEvent(s)?.let { return it }
        if (s.category == PostCategories.IDLE && s.enableIdlePrice) {
            val price = s.idlePrice.toDoubleOrNull()
            if (price != null && price < 0) return R.string.post_price_invalid
            if ((price == null || price <= 0.0) && !s.idleNegotiable) return R.string.post_need_price
        }
        if (s.category == PostCategories.HELP && s.helpReward) {
            val reward = s.helpRewardAmount.toDoubleOrNull()
            if (reward == null || reward <= 0.0) return R.string.post_need_reward
        }
        if (s.hasUploadingImages) return R.string.post_images_uploading_wait
        if (s.images.any { it.failed }) return R.string.post_images_failed
        return null
    }

    @StringRes
    private fun validateEvent(s: PostCreateUiState): Int? {
        if (s.enableSchedule && s.startAt > 0 && s.endAt > 0 && s.endAt < s.startAt) {
            return R.string.post_end_before_start
        }
        if ((s.needSignup || s.needCheckIn) && (!s.enableSchedule || s.endAt <= 0)) {
            return R.string.post_need_end_for_signup
        }
        if (s.enableSchedule && s.startAt <= 0 && s.endAt <= 0) return R.string.post_need_schedule
        if (s.needSignup && s.limitParticipants) {
            val max = s.maxParticipants.toIntOrNull()
            if (max == null || max <= 0) return R.string.post_need_max
        }
        if (s.needSignup && s.signupDeadlineAt > 0) {
            if (s.enableSchedule && s.startAt > 0 && s.signupDeadlineAt > s.startAt) {
                return R.string.post_deadline_after_start
            }
            if (PostSignupRules.isSignupDeadlinePassed(s.signupDeadlineAt)) return R.string.post_deadline_passed
        }
        if (s.enableFee) {
            val fee = s.feeAmount.toDoubleOrNull()
            if (fee == null || fee <= 0.0) return R.string.post_need_fee
        }
        return null
    }

    fun toDraft(s: PostCreateUiState): PostDraft {
        var startAt = 0L
        var endAt = 0L
        var extras: PostExtras? = null

        when (s.category) {
            PostCategories.EVENT -> {
                if (s.enableSchedule) {
                    startAt = s.startAt
                    endAt = s.endAt
                }
                extras = PostExtras(
                    needSignup = s.needSignup,
                    needCheckIn = s.needSignup && s.needCheckIn,
                    maxParticipants = if (s.needSignup && s.limitParticipants) {
                        s.maxParticipants.toIntOrNull()?.coerceAtLeast(0) ?: 0
                    } else {
                        0
                    },
                    allowWaitlist = s.needSignup && s.limitParticipants && s.allowWaitlist,
                    signupDeadlineAt = if (s.needSignup) s.signupDeadlineAt else 0,
                    feeType = when {
                        !s.enableFee -> FEE_TYPE_NONE
                        s.feeKind == FeeKind.Reward -> FEE_TYPE_REWARD
                        else -> FEE_TYPE_CHARGE
                    },
                    amount = if (s.enableFee) s.feeAmount.toDoubleOrNull() ?: 0.0 else 0.0,
                    feeDescription = if (s.enableFee) s.feeDescription.trim() else "",
                )
            }

            PostCategories.IDLE -> {
                if (s.enableIdlePrice) {
                    extras = PostExtras(
                        amount = s.idlePrice.toDoubleOrNull() ?: 0.0,
                        negotiable = s.idleNegotiable,
                    )
                }
            }

            PostCategories.HELP -> {
                if (s.helpReward) {
                    extras = PostExtras(
                        feeType = FEE_TYPE_REWARD,
                        amount = s.helpRewardAmount.toDoubleOrNull() ?: 0.0,
                        feeDescription = s.helpRewardDesc.trim(),
                    )
                }
            }
        }

        return PostDraft(
            id = s.editPostId,
            roomId = s.roomId,
            category = s.category,
            pinned = s.isRoomPost && s.pinned,
            title = s.title.trim(),
            content = s.content.trim(),
            images = s.imageMediaIds,
            placeName = s.placeName.trim(),
            latitude = s.latitude,
            longitude = s.longitude,
            startAt = startAt,
            endAt = endAt,
            extras = extras,
        )
    }

    fun fromPost(s: PostCreateUiState, d: PostItem): PostCreateUiState {
        val extras = d.extras
        val amount = extras?.amount ?: 0.0
        val enableSchedule = d.startAt > 0 || d.endAt > 0
        val enableFee = (extras?.feeType ?: 0) in FEE_TYPE_CHARGE..FEE_TYPE_REWARD && amount > 0
        val enableIdlePrice = d.category == PostCategories.IDLE && (amount > 0 || extras?.negotiable == true)
        val helpReward = d.category == PostCategories.HELP &&
            extras != null &&
            extras.feeType == FEE_TYPE_REWARD &&
            (extras.amount > 0 || extras.feeDescription.isNotBlank())
        return s.copy(
            pageLoading = false,
            roomId = d.roomId,
            category = d.category,
            pinned = d.pinned,
            title = d.title,
            content = d.content,
            placeName = d.placeName,
            latitude = d.latitude,
            longitude = d.longitude,
            locationReady = d.latitude != 0.0 || d.longitude != 0.0 || d.placeName.isNotBlank(),
            enableSchedule = enableSchedule,
            startAt = d.startAt,
            endAt = d.endAt,
            enableFee = enableFee && d.category == PostCategories.EVENT,
            feeKind = if (extras?.feeType == FEE_TYPE_REWARD) FeeKind.Reward else FeeKind.Charge,
            feeAmount = if (enableFee) amountField(amount) else "",
            feeDescription = extras?.feeDescription.orEmpty(),
            needSignup = extras?.needSignup == true,
            needCheckIn = extras?.needSignup == true && extras.needCheckIn,
            limitParticipants = (extras?.maxParticipants ?: 0) > 0,
            maxParticipants = extras?.maxParticipants?.takeIf { it > 0 }?.toString().orEmpty(),
            allowWaitlist = extras?.allowWaitlist == true,
            signupDeadlineAt = extras?.signupDeadlineAt ?: 0,
            enableIdlePrice = enableIdlePrice,
            idlePrice = if (enableIdlePrice) amountField(amount) else "",
            idleNegotiable = extras?.negotiable != false,
            helpReward = helpReward,
            helpRewardAmount = if (helpReward) amountField(amount) else "",
            helpRewardDesc = if (helpReward) extras?.feeDescription.orEmpty() else "",
            images = d.images.mapIndexed { index, url ->
                PostImageItem(
                    id = UUID.randomUUID().toString(),
                    localUri = url,
                    url = url,
                    mediaId = d.imageIds.getOrNull(index) ?: 0L,
                    uploading = false,
                )
            },
        )
    }

    fun filterNumber(value: String): String {
        var dot = false
        return buildString {
            for (c in value) {
                when {
                    c.isDigit() -> append(c)

                    c == '.' && !dot -> {
                        append(c)
                        dot = true
                    }
                }
            }
        }
    }

    fun filterDigits(value: String): String = value.filter { it.isDigit() }

    private fun amountField(v: Double): String = if (v == 0.0) "" else MoneyFormat.formatAmount(v)
}
