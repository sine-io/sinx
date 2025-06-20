import { Admin, Resource, CustomRoutes } from 'react-admin';
import { Route } from "react-router-dom";
import PlayCircleOutlineIcon from '@mui/icons-material/PlayCircleOutline';
import { createHashHistory } from "history";

import jobs from './jobs';
import { BusyList } from './executions/BusyList';
import { Layout } from './layout';
import dataProvider from './dataProvider';
import authProvider from './authProvider';
import Dashboard from './dashboard';
import Settings from './settings/Settings';
import LoginPage from './LoginPage';

declare global {
    interface Window {
        SINX_API_URL: string;
        SINX_LEADER: string;
        SINX_UNTRIGGERED_JOBS: string;
        SINX_FAILED_JOBS: string;
        SINX_SUCCESSFUL_JOBS: string;
        SINX_TOTAL_JOBS: string;
        SINX_ACL_ENABLED: boolean;
    }
}

const history = createHashHistory();
 
export const App = () => <Admin
    dashboard={Dashboard}
    loginPage={LoginPage}
    authProvider={window.SINX_ACL_ENABLED ? authProvider : undefined}
    dataProvider={dataProvider}
    layout={Layout}
>

    <Resource name="jobs" {...jobs} />
    <Resource name="busy" options={{ label: 'Busy' }} list={BusyList} icon={PlayCircleOutlineIcon} />
    <Resource name="executions" />
    <Resource name="members" />
    <CustomRoutes>
        <Route path="/settings" element={<Settings />} />
    </CustomRoutes>
</Admin>;
