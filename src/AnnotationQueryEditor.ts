export class AnnotationQueryEditor {
  static templateUrl = 'partials/annotations.editor.html';

  annotation: any;

  constructor() {
    this.annotation.spaceName = this.annotation.spaceName || '';
    this.annotation.projectName = this.annotation.projectName || '';
    this.annotation.environmentName = this.annotation.environmentName || '';
  }
}
