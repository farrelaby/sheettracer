export type Selection =
  | { type: "node"; id: string }
  | { type: "edge"; id: string };

class SelectionState {
  value = $state<Selection | null>(null);

  set(value: Selection | null) {
    this.value = value;
  }

  isNull() {
    return this.value === null;
  }
}

export const selectionStore = new SelectionState();
