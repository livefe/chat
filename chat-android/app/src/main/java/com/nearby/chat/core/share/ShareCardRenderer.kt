package com.nearby.chat.core.share

import android.graphics.Bitmap
import android.graphics.Canvas
import android.graphics.Color
import android.graphics.LinearGradient
import android.graphics.Paint
import android.graphics.RectF
import android.graphics.Shader
import android.graphics.Typeface
import android.text.Layout
import android.text.StaticLayout
import android.text.TextPaint
import android.text.TextUtils

internal object ShareCardRenderer {
    const val QR_SIZE = 200

    private const val CARD_W = 1080
    private const val COVER_H = 810
    private const val PAD = 48
    private const val FOOTER_H = 280

    fun render(
        title: String,
        content: String,
        cover: Bitmap?,
        qr: Bitmap,
        appLabel: String,
        scanTip: String,
    ): Bitmap {
        val cardH = COVER_H + FOOTER_H
        val bitmap = Bitmap.createBitmap(CARD_W, cardH, Bitmap.Config.ARGB_8888)
        val canvas = Canvas(bitmap)
        canvas.drawColor(Color.WHITE)

        if (cover != null) {
            drawCover(canvas, cover, CARD_W, COVER_H)
            cover.recycle()
            val shade = Paint(Paint.ANTI_ALIAS_FLAG).apply {
                shader = LinearGradient(
                    0f,
                    COVER_H - 120f,
                    0f,
                    COVER_H.toFloat(),
                    0x00000000,
                    0x33000000,
                    Shader.TileMode.CLAMP,
                )
            }
            canvas.drawRect(0f, COVER_H - 120f, CARD_W.toFloat(), COVER_H.toFloat(), shade)
        } else {
            drawTextCover(canvas, content)
        }

        val footerTop = COVER_H.toFloat()
        canvas.drawRect(
            0f,
            footerTop,
            CARD_W.toFloat(),
            cardH.toFloat(),
            Paint().apply {
                color = Color.WHITE
            },
        )

        val textRight = CARD_W - PAD - QR_SIZE - 40
        val textWidth = (textRight - PAD).coerceAtLeast(1)
        val titlePaint = TextPaint(Paint.ANTI_ALIAS_FLAG).apply {
            color = 0xFF0F0F0F.toInt()
            textSize = 48f
            typeface = Typeface.create(Typeface.DEFAULT, Typeface.BOLD)
        }
        val tipPaint = TextPaint(Paint.ANTI_ALIAS_FLAG).apply {
            color = 0xFF888888.toInt()
            textSize = 26f
        }
        val brandPaint = TextPaint(Paint.ANTI_ALIAS_FLAG).apply {
            color = 0xFF4C8DFF.toInt()
            textSize = 28f
            typeface = Typeface.create(Typeface.DEFAULT, Typeface.BOLD)
        }

        val titleLayout = buildLayout(title, titlePaint, textWidth, maxLines = 2)
        val y = footerTop + 52f
        canvas.save()
        canvas.translate(PAD.toFloat(), y)
        titleLayout.draw(canvas)
        canvas.restore()

        val brandY = cardH - PAD - 8f
        canvas.drawText(appLabel, PAD.toFloat(), brandY, brandPaint)
        canvas.drawText(scanTip, PAD + brandPaint.measureText(appLabel) + 16f, brandY, tipPaint)

        val qrLeft = (CARD_W - PAD - QR_SIZE).toFloat()
        val qrTop = footerTop + (FOOTER_H - QR_SIZE) / 2f
        val qrFrame = RectF(qrLeft - 14f, qrTop - 14f, qrLeft + QR_SIZE + 14f, qrTop + QR_SIZE + 14f)
        val frameFill = Paint(Paint.ANTI_ALIAS_FLAG).apply { color = Color.WHITE }
        val frameStroke = Paint(Paint.ANTI_ALIAS_FLAG).apply {
            style = Paint.Style.STROKE
            strokeWidth = 2f
            color = 0x1A000000
        }
        canvas.drawRoundRect(qrFrame, 18f, 18f, frameFill)
        canvas.drawRoundRect(qrFrame, 18f, 18f, frameStroke)
        canvas.drawBitmap(qr, qrLeft, qrTop, null)

        return bitmap
    }

    private fun drawCover(canvas: Canvas, cover: Bitmap, w: Int, h: Int) {
        val scale = maxOf(w.toFloat() / cover.width, h.toFloat() / cover.height)
        val dw = cover.width * scale
        val dh = cover.height * scale
        val left = (w - dw) / 2f
        val top = (h - dh) / 2f
        val paint = Paint(Paint.ANTI_ALIAS_FLAG or Paint.FILTER_BITMAP_FLAG)
        canvas.save()
        canvas.clipRect(0, 0, w, h)
        canvas.drawBitmap(cover, null, RectF(left, top, left + dw, top + dh), paint)
        canvas.restore()
    }

    private fun drawTextCover(canvas: Canvas, content: String) {
        val bg = Paint(Paint.ANTI_ALIAS_FLAG).apply {
            shader = LinearGradient(
                0f,
                0f,
                CARD_W.toFloat(),
                COVER_H.toFloat(),
                intArrayOf(0xFFF5F5F5.toInt(), 0xFFE5E5E5.toInt()),
                floatArrayOf(0f, 1f),
                Shader.TileMode.CLAMP,
            )
        }
        canvas.drawRect(0f, 0f, CARD_W.toFloat(), COVER_H.toFloat(), bg)

        val body = content.trim().replace(Regex("\\s+"), " ")
        if (body.isBlank()) return

        val bodyPaint = TextPaint(Paint.ANTI_ALIAS_FLAG).apply {
            color = 0xFF555555.toInt()
            textSize = 46f
        }
        val maxWidth = CARD_W - PAD * 2
        val centered = StaticLayout.Builder
            .obtain(body, 0, body.length, bodyPaint, maxWidth)
            .setAlignment(Layout.Alignment.ALIGN_CENTER)
            .setLineSpacing(8f, 1.1f)
            .setIncludePad(false)
            .setMaxLines(6)
            .setEllipsize(TextUtils.TruncateAt.END)
            .build()
        val top = (COVER_H - centered.height) / 2f
        canvas.save()
        canvas.translate(PAD.toFloat(), top)
        centered.draw(canvas)
        canvas.restore()
    }

    private fun buildLayout(text: String, paint: TextPaint, width: Int, maxLines: Int): StaticLayout = StaticLayout.Builder
        .obtain(text, 0, text.length, paint, width)
        .setAlignment(Layout.Alignment.ALIGN_NORMAL)
        .setLineSpacing(4f, 1.1f)
        .setIncludePad(false)
        .setMaxLines(maxLines)
        .setEllipsize(TextUtils.TruncateAt.END)
        .build()
}
