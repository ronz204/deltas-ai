import assert from "node:assert";
import { test } from "node:test";

test("INV-02: fails", () => {
  assert.fail("this claim is violated");
});
