package com.nearby.chat.data.export

import java.io.ByteArrayOutputStream
import java.util.Locale
import java.util.zip.ZipEntry
import java.util.zip.ZipOutputStream
import kotlin.math.max
import kotlin.math.min

object XlsxWriter {
    const val MIME_XLSX = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
    private const val DEFAULT_ROW_HEIGHT = 22.0
    private const val HEADER_ROW_HEIGHT = 24.0
    private const val DEFAULT_FONT_NAME = "微软雅黑"

    fun build(
        sheetName: String,
        headers: List<String>,
        rows: List<List<String>>,
    ): ByteArray {
        val table = listOf(headers) + rows
        val shared = LinkedHashMap<String, Int>()
        fun sid(text: String): Int = shared.getOrPut(text) { shared.size }

        val cellRefs = table.map { row ->
            row.map { cell -> sid(cell) }
        }
        val widths = headers.indices.map { i ->
            columnWidthCh(headers[i], rows.map { it[i] })
        }

        val sharedXml = buildSharedStrings(shared.keys.toList())
        val sheetXml = buildSheet(cellRefs, widths, headers.size)

        return ByteArrayOutputStream().use { bos ->
            ZipOutputStream(bos).use { zip ->
                putZip(zip, "[Content_Types].xml", CONTENT_TYPES)
                putZip(zip, "_rels/.rels", ROOT_RELS)
                putZip(zip, "xl/workbook.xml", buildWorkbook(sheetName))
                putZip(zip, "xl/_rels/workbook.xml.rels", WORKBOOK_RELS)
                putZip(zip, "xl/styles.xml", STYLES_XML)
                putZip(zip, "xl/sharedStrings.xml", sharedXml)
                putZip(zip, "xl/worksheets/sheet1.xml", sheetXml)
            }
            bos.toByteArray()
        }
    }

    private fun buildWorkbook(sheetName: String): String = buildString {
        append("""<?xml version="1.0" encoding="UTF-8" standalone="yes"?>""")
        append('\n')
        append(
            """<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" """ +
                """xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">""",
        )
        append('\n')
        append("  <sheets>\n")
        append("""    <sheet name="${escapeXml(sheetName)}" sheetId="1" r:id="rId1"/>""")
        append('\n')
        append("  </sheets>\n")
        append("</workbook>")
    }

    private fun buildSharedStrings(strings: List<String>): String = buildString {
        append("""<?xml version="1.0" encoding="UTF-8" standalone="yes"?>""")
        append("""<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" """)
        append("""count="${strings.size}" uniqueCount="${strings.size}">""")
        strings.forEach { s ->
            append("<si>")
            if (s.isEmpty()) {
                append("<t></t>")
            } else if (s.first().isWhitespace() || s.last().isWhitespace()) {
                append("""<t xml:space="preserve">${escapeXml(s)}</t>""")
            } else {
                append("<t>${escapeXml(s)}</t>")
            }
            append("</si>")
        }
        append("</sst>")
    }

    private fun buildSheet(cellRefs: List<List<Int>>, widths: List<Double>, colCount: Int): String = buildString {
        val lastCol = colName(colCount - 1)
        val lastRow = cellRefs.size
        append("""<?xml version="1.0" encoding="UTF-8" standalone="yes"?>""")
        append("""<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">""")
        append("""<dimension ref="A1:$lastCol$lastRow"/>""")
        append("<cols>")
        widths.forEachIndexed { i, w ->
            val col = i + 1
            append(
                """<col min="$col" max="$col" width="${"%.2f".format(Locale.US, w)}" customWidth="1"/>""",
            )
        }
        append("</cols>")
        append("""<sheetData>""")
        cellRefs.forEachIndexed { rIdx, row ->
            val r = rIdx + 1
            val ht = if (rIdx == 0) HEADER_ROW_HEIGHT else DEFAULT_ROW_HEIGHT
            val style = 0
            append("""<row r="$r" ht="${"%.1f".format(Locale.US, ht)}" customHeight="1">""")
            row.forEachIndexed { cIdx, sid ->
                val ref = "${colName(cIdx)}$r"
                append("""<c r="$ref" t="s" s="$style"><v>$sid</v></c>""")
            }
            append("</row>")
        }
        append("</sheetData>")
        append("</worksheet>")
    }

    private fun columnWidthCh(header: String, values: List<String>): Double {
        var maxChars = displayWidth(header)
        values.forEach { maxChars = max(maxChars, displayWidth(it)) }
        return min(48.0, max(10.0, maxChars + 2.0))
    }

    private fun displayWidth(text: String): Int {
        var w = 0
        for (ch in text) {
            w += if (ch.code > 0xFF) 2 else 1
        }
        return w
    }

    private fun colName(index: Int): String {
        var n = index
        val sb = StringBuilder()
        while (n >= 0) {
            sb.insert(0, ('A'.code + n % 26).toChar())
            n = n / 26 - 1
        }
        return sb.toString()
    }

    private fun putZip(zip: ZipOutputStream, path: String, content: String) {
        zip.putNextEntry(ZipEntry(path))
        zip.write(content.toByteArray(Charsets.UTF_8))
        zip.closeEntry()
    }

    private fun escapeXml(value: String): String = buildString(value.length + 8) {
        for (ch in value) {
            when (ch) {
                '&' -> append("&amp;")
                '<' -> append("&lt;")
                '>' -> append("&gt;")
                '"' -> append("&quot;")
                '\'' -> append("&apos;")
                in '\u0000'..'\u0008', '\u000B', '\u000C', in '\u000E'..'\u001F' -> Unit
                else -> append(ch)
            }
        }
    }

    private const val CONTENT_TYPES = """<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
  <Override PartName="/xl/sharedStrings.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sharedStrings+xml"/>
  <Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>
</Types>"""

    private const val ROOT_RELS = """<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>"""

    private const val WORKBOOK_RELS = """<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings" Target="sharedStrings.xml"/>
</Relationships>"""

    private const val STYLES_XML = """<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <fonts count="1">
    <font><sz val="11"/><name val="$DEFAULT_FONT_NAME"/></font>
  </fonts>
  <fills count="2">
    <fill><patternFill patternType="none"/></fill>
    <fill><patternFill patternType="gray125"/></fill>
  </fills>
  <borders count="1"><border/></borders>
  <cellStyleXfs count="1"><xf/></cellStyleXfs>
  <cellXfs count="1">
    <xf fontId="0" fillId="0" borderId="0" xfId="0"/>
  </cellXfs>
</styleSheet>"""
}
