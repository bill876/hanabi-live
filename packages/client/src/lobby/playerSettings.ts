// The "Settings" nav button.

import { parseIntSafe } from "@hanabi/data";
import globals from "../globals";
import * as notifications from "../notifications";
import { isKeyOf } from "../utils";

export function init(): void {
  $("#settings-volume-slider").change(function settingsVolumeSliderChange(
    this: HTMLElement,
  ) {
    const element = $(this);
    const volumeString = element.val();
    if (typeof volumeString !== "string") {
      throw new Error(
        'The value of the "settings-volume-slider" element is not a string.',
      );
    }
    const volume = parseIntSafe(volumeString);
    globals.settings.volume = volume;
    $("#settings-volume-slider-value").html(`${volume}%`);
    globals.conn!.send("setting", {
      name: "volume",
      setting: volumeString, // The server expects all setting values as strings
    });
  });

  $("#settings-volume-test").click(() => {
    const audio = new Audio("/public/sounds/turn_us.mp3");
    const element = $("#settings-volume-slider");
    const volumeString = element.val();
    if (typeof volumeString !== "string") {
      throw new Error(
        'The value of the "settings-volume-slider" element is not a string.',
      );
    }
    const volume = parseIntSafe(volumeString);
    audio.volume = volume / 100;
    audio.play().catch((err) => {
      console.error("Failed to play the test sound:", err);
    });
  });
}

export function setPlayerSettings(): void {
  // The server has delivered to us a list of all of our settings. Check the checkboxes for the
  // settings that we have enabled (and adjust the volume slider).
  // eslint-disable-next-line isaacscript/no-object-any
  for (const [setting, value] of Object.entries(globals.settings)) {
    if (setting.startsWith("createTable")) {
      // Settings for the "Create Game" nav button are handled when the user clicks on it.
      continue;
    } else if (setting === "volume") {
      if (typeof value !== "number") {
        throw new Error("The volume setting is not stored as a number.");
      }
      $("#settings-volume-slider").val(value);
      $("#settings-volume-slider-value").html(`${value}%`);
    } else {
      const element = $(`#${setting}`);
      if (typeof value !== "boolean") {
        throw new Error("The volume setting is not stored as a string.");
      }
      element.prop("checked", value);
      element.change(changeSetting);
    }
  }
}

function changeSetting(this: HTMLElement) {
  const element = $(this);
  const settingName = element.attr("id");
  if (settingName === undefined || settingName === "") {
    throw new Error(
      'Failed to get the ID of the element in the "changeSetting()" function.',
    );
  }
  if (!isKeyOf(settingName, globals.settings)) {
    throw new Error(
      `The setting of ${settingName} does not exist in the Settings class.`,
    );
  }
  const setting = globals.settings[settingName];

  const checked = element.is(":checked");
  if (setting === checked) {
    return;
  }

  // Write the new value to our local variable. We must cast the settings to any since this
  // assignment violates type safety.
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  (globals.settings[settingName] as any) = checked;

  // Send the new value to the server.
  globals.conn!.send("setting", {
    name: settingName,
    setting: checked.toString(), // The server expects all setting values as strings.
  });

  if (settingName === "desktopNotification" && checked) {
    notifications.test();
  }
}
