import unittest
from PIL import Image, ImageDraw
from desktop import measure


class RegionTests(unittest.TestCase):
    def test_colour_outside_component_cannot_mask_missing_component(self):
        image = Image.new("RGB", (100, 100), "red")
        ImageDraw.Draw(image).rectangle((0, 0, 19, 19), fill="black")
        metrics, _ = measure(image, [("module", (0, 0, 20, 20), (255, 0, 0), .5)])
        self.assertFalse(metrics[0]["passed"])
        self.assertEqual(metrics[0]["value"], 0)

    def test_region_reports_measured_fraction(self):
        image = Image.new("RGB", (20, 20), "black")
        ImageDraw.Draw(image).rectangle((0, 0, 9, 19), fill="red")
        metrics, _ = measure(image, [("module", (0, 0, 20, 20), (255, 0, 0), .4)])
        self.assertTrue(metrics[0]["passed"])
        self.assertEqual(metrics[0]["value"], .5)
