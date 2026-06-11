import {useCallback, useEffect, useRef, useState, type CSSProperties} from 'react';
import './App.css';
import {Quit, WindowIsMaximised, WindowMinimise, WindowToggleMaximise} from '../wailsjs/runtime/runtime';

const dragRegionStyle = {'--wails-draggable': 'drag'} as CSSProperties;

function App() {
    const [isMaximised, setIsMaximised] = useState(false);
    const resizeTimer = useRef<number>();

    const syncMaximisedState = useCallback(() => {
        WindowIsMaximised().then(setIsMaximised).catch(() => setIsMaximised(false));
    }, []);

    const toggleMaximise = () => {
        WindowToggleMaximise();
        window.setTimeout(syncMaximisedState, 80);
    };

    useEffect(() => {
        syncMaximisedState();

        const handleResize = () => {
            window.clearTimeout(resizeTimer.current);
            resizeTimer.current = window.setTimeout(syncMaximisedState, 160);
        };

        window.addEventListener('resize', handleResize);

        return () => {
            window.clearTimeout(resizeTimer.current);
            window.removeEventListener('resize', handleResize);
        };
    }, [syncMaximisedState]);

    return (
        <main id="App" className={isMaximised ? 'is-maximised' : undefined}>
            <section className="window-shell">
                <header className="titlebar">
                    <div className="drag-region" style={dragRegionStyle}>
                        <span className="app-dot"/>
                        <span className="title">wails-rounded-window</span>
                    </div>
                    <div className="window-actions" aria-label="Window controls">
                        <button className="window-button" type="button" aria-label="Minimise window" onClick={WindowMinimise}>
                            -
                        </button>
                        <button className="window-button" type="button" aria-label="Maximise window" onClick={toggleMaximise}>
                            []
                        </button>
                        <button className="window-button close" type="button" aria-label="Close app" onClick={Quit}>
                            x
                        </button>
                    </div>
                </header>

                <div className="content">
                    <div className="status-panel">
                        <p className="eyebrow">Wails2 frameless demo</p>
                        <h1>24px rounded window</h1>
                        <p className="description">
                            Drag the custom titlebar, resize the edges, and use the window controls to verify the
                            frameless shell.
                        </p>
                        <div className="checks">
                            <span>Frameless</span>
                            <span>Transparent WebView</span>
                            <span>Custom controls</span>
                        </div>
                    </div>
                </div>
            </section>
        </main>
    );
}

export default App;
