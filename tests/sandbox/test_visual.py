import unittest
from PIL import Image, ImageDraw
from visual import analyse, contrast, missing_text


class VisualTests(unittest.TestCase):
    def fixture(self):
        image = Image.new("RGB", (1280, 900), (48, 52, 59))
        draw = ImageDraw.Draw(image)
        draw.rectangle((780, 0, 1279, 899), fill=(30, 30, 46))
        draw.rectangle((810, 145, 1250, 265), fill=(43, 43, 63))
        draw.rectangle((810, 148, 1249, 197), outline=(243, 139, 168), width=2)
        # Synthetic glyph-like marks isolate colour/layout gates from OCR.
        for x in range(820, 1200, 8):
            draw.rectangle((x, 160, x + 2, 168), fill=(205, 214, 244))
            draw.rectangle((x, 30, x + 2, 38), fill=(205, 214, 244))
        return image

    def failures(self, image):
        report, _ = analyse(image, "dark", (43, 43, 63), True, True)
        return {m["name"] for m in report["metrics"] if not m["passed"]}

    def test_healthy_fixture(self):
        self.assertEqual(self.failures(self.fixture()), set())

    def test_blank_screen_rejected(self):
        self.assertIn("notification_surface_fraction", self.failures(Image.new("RGB", (1280, 900))))

    def test_hidden_notification_text_rejected(self):
        image = self.fixture()
        pixels = image.load()
        for y in range(145, 265):
            for x in range(810, 1250):
                if pixels[x, y] == (205, 214, 244):
                    pixels[x, y] = (43, 43, 63)
        self.assertIn("notification_text_fraction", self.failures(image))

    def test_missing_panel_rejected_even_with_healthy_notifications(self):
        image = self.fixture()
        ImageDraw.Draw(image).rectangle((790, 300, 1270, 880), fill=(48, 52, 59))
        self.assertIn("panel_background_fraction", self.failures(image))

    def test_missing_critical_border_rejected(self):
        image = self.fixture()
        ImageDraw.Draw(image).rectangle((810, 148, 1249, 197), outline=(43, 43, 63), width=2)
        self.assertIn("critical_border_top_fraction", self.failures(image))

    def test_interior_danger_colour_cannot_replace_border(self):
        image = self.fixture()
        draw = ImageDraw.Draw(image)
        draw.rectangle((810, 148, 1249, 197), outline=(43, 43, 63), width=2)
        draw.rectangle((830, 175, 1200, 185), fill=(243, 139, 168))
        self.assertIn("critical_border_top_fraction", self.failures(image))
        self.assertIn("critical_border_left_fraction", self.failures(image))

    def test_offscreen_content_rejected(self):
        image = Image.new("RGB", (1280, 900), (48, 52, 59))
        image.paste(self.fixture().crop((780, 0, 1280, 900)), (1100, 0))
        self.assertTrue(self.failures(image))

    def test_contrast_reference_values(self):
        self.assertAlmostEqual(contrast((0, 0, 0), (255, 255, 255)), 21)
        self.assertAlmostEqual(contrast((100, 100, 100), (100, 100, 100)), 1)

    def test_ocr_requires_every_phrase(self):
        self.assertEqual(missing_text("Normal\nnotification", ["normal notification"]), [])
        self.assertEqual(missing_text("Normal notification", ["Normal notification", "Readable body text"]),
                         ["Readable body text"])


if __name__ == "__main__":
    unittest.main()
