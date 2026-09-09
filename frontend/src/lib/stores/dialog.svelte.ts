class DialogTracker {
  openCount = $state(0);

  push() {
    this.openCount += 1;
  }

  pop() {
    this.openCount = Math.max(0, this.openCount - 1);
  }

  get isOpen() {
    return this.openCount > 0;
  }
}

export const dialogTracker = new DialogTracker();
