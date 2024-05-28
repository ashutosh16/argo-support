import {App, Flyout} from './app';



const TITLE = 'Unified Visualizer';
const ID = 'uv';

((window: any) => {
    window?.extensionsAPI?.registerStatusPanelExtension(App, TITLE, ID, Flyout);
})(window);