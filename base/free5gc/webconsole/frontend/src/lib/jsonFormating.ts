export interface MultipleDeleteSubscriberData {
  ueId: string;
  plmnID: string;
}

export function formatMultipleDeleteSubscriberToJson(subscribers: MultipleDeleteSubscriberData[]) {
  return subscribers.map(sub => ({
    ueId: sub.ueId,
    plmnID: sub.plmnID
  }));
}

export interface MultipleDeleteProfileData {
  profileName: string;
}

export function formatMultipleDeleteProfileToJson(profiles: MultipleDeleteProfileData[]) {
  return profiles.map(profile => ({
    profileName: profile.profileName
  }));
}