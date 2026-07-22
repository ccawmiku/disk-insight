import { expect, test } from "@playwright/test";

test("renders the complete analytics dashboard", async ({ page }) => {
  await page.goto("/");
  await expect(page).toHaveTitle("Disk Insight");
  await expect(page.getByText("文件总大小", { exact: true })).toBeVisible();
  await expect(page.getByText("文件大小分布", { exact: true })).toBeVisible();
  await expect(page.getByText("累计百分比", { exact: true })).toBeVisible();
  await expect(page.getByText("文件类别", { exact: true })).toBeVisible();
  await expect.poll(() => page.locator("canvas").count()).toBeGreaterThanOrEqual(5);
  await expect(page.locator(".waffle > span")).toHaveCount(200);
});

test("supports navigation and category filters", async ({ page }, testInfo) => {
  await page.goto("/");
  if (testInfo.project.name.startsWith("mobile")) {
    await page.getByRole("button", { name: "打开菜单", exact: true }).click();
  }
  const filesButton = page.getByRole("button", { name: "文件浏览", exact: true });
  await expect(filesButton).toBeVisible();
  await filesButton.click();
  await expect(page.getByText("最大的 100 个文件", { exact: true })).toBeVisible();
  await expect(page.locator("tbody tr")).not.toHaveCount(0);
});

test("keeps the mobile layout inside the viewport", async ({ page }, testInfo) => {
  test.skip(!testInfo.project.name.startsWith("mobile"), "mobile-only assertion");
  await page.goto("/");
  await expect(page.getByRole("button", { name: "打开菜单", exact: true })).toBeVisible();
  const overflow = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
  expect(overflow).toBeLessThanOrEqual(1);
});
