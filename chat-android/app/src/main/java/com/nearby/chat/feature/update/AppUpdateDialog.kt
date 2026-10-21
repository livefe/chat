package com.nearby.chat.feature.update

import android.content.Intent
import android.net.Uri
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Text
import androidx.compose.material3.VerticalDivider
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.DialogProperties
import com.nearby.chat.R
import com.nearby.chat.core.model.AppUpdateInfo
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppShapes
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType

private val DialogHorizontalMargin = 40.dp
private val ActionRowHeight = 48.dp
private val HairlineThickness = 0.5.dp

@Composable
fun AppUpdateDialog(
    data: AppUpdateInfo,
    onDismissSoft: () -> Unit,
    onDownloadUnavailable: () -> Unit,
) {
    val context = LocalContext.current
    val force = data.forceUpdate
    Dialog(
        onDismissRequest = { if (!force) onDismissSoft() },
        properties = DialogProperties(
            dismissOnBackPress = !force,
            dismissOnClickOutside = !force,
            usePlatformDefaultWidth = false,
        ),
    ) {
        Column(
            modifier = Modifier
                .padding(horizontal = DialogHorizontalMargin)
                .fillMaxWidth()
                .clip(AppShapes.CardLg)
                .background(AppColors.CardBg),
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(start = AppSpacing.Xl, end = AppSpacing.Xl, top = AppSpacing.Xxl, bottom = AppSpacing.Xl),
            ) {
                Text(
                    data.title.ifBlank { stringResource(if (force) R.string.update_required else R.string.update_found) },
                    fontSize = AppType.PageTitle,
                    fontWeight = FontWeight.SemiBold,
                    color = AppColors.TextPrimary,
                    modifier = Modifier.fillMaxWidth(),
                )
                val message = data.message.ifBlank { stringResource(R.string.update_default_message) }
                val versionLine = if (data.latestVersion.isNotBlank()) {
                    stringResource(R.string.update_latest_label, data.latestVersion)
                } else {
                    ""
                }
                val body = if (versionLine.isEmpty()) message else "$message\n\n$versionLine"
                Text(
                    body,
                    fontSize = AppType.Secondary,
                    color = AppColors.TextSecondary,
                    lineHeight = 22.sp,
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = AppSpacing.S),
                )
            }
            HorizontalDivider(color = AppColors.BorderSoft, thickness = HairlineThickness)
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(ActionRowHeight),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                if (!force) {
                    Text(
                        stringResource(R.string.update_later),
                        fontSize = AppType.Body,
                        color = AppColors.TextSecondary,
                        textAlign = TextAlign.Center,
                        modifier = Modifier
                            .weight(1f)
                            .clickable(onClick = onDismissSoft)
                            .padding(vertical = AppSpacing.M),
                    )
                    VerticalDivider(color = AppColors.BorderSoft, thickness = HairlineThickness)
                }
                Text(
                    stringResource(R.string.update_now),
                    fontSize = AppType.Body,
                    fontWeight = FontWeight.Medium,
                    color = AppColors.Brand,
                    textAlign = TextAlign.Center,
                    modifier = Modifier
                        .weight(1f)
                        .clickable {
                            val url = data.downloadUrl.trim()
                            if (url.isNotEmpty()) {
                                runCatching {
                                    context.startActivity(Intent(Intent.ACTION_VIEW, Uri.parse(url)))
                                }
                            } else {
                                onDownloadUnavailable()
                            }
                        }
                        .padding(vertical = AppSpacing.M),
                )
            }
        }
    }
}
