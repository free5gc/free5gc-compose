import React, { useState } from "react";
import { useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";

import axios from "../../axios";

import Dashboard from "../../Dashboard";
import { Button, Grid } from "@mui/material";
import { ProfileFormProvider, useProfileForm } from "../../hooks/profile-form";
import ProfileFormBasic from "./ProfileFormBasic";
import ProfileFormUeAmbr from "./ProfileFormUeAmbr";
import ProfileFormSessions from "./ProfileFormSessions";
import { ProfileMapperImpl, FlowsMapperImpl } from "../../lib/dtos/profile";
import { validateProfile } from "../../lib/validator/profileValidator";

function FormHOC(Component: React.ComponentType<any>) {
    return function (props: any) {
        return (
            <ProfileFormProvider>
                <Component {...props} />
            </ProfileFormProvider>
        );
    };
}

export default FormHOC(ProfileCreate);

function ProfileCreate() {
  const { profileName } = useParams<{ profileName: string }>();

  const isNewProfile = profileName === undefined;
  const navigation = useNavigate();
  const [loading, setLoading] = useState(false);

  const { handleSubmit, getValues, reset } = useProfileForm();

  if (!isNewProfile) {
    useEffect(() => {
      setLoading(true);

      axios
        .get("/api/profile/" + profileName)
        .then((res) => {
          const profileMapper = new ProfileMapperImpl(new FlowsMapperImpl());
          const profile = profileMapper.mapFromProfile(res.data);
          reset(profile);
        })
        .finally(() => {
          setLoading(false);
        });
    }, [profileName]);
  }

  if (loading) {
    return <div>Loading...</div>;
  }

  const onCreate = () => {
    console.log("trace: onCreate");

    const data = getValues();

    if (data.SnssaiConfigurations.length === 0) {
      alert("Please add at least one S-NSSAI");
      return;
    }

    const profileMapper = new ProfileMapperImpl(new FlowsMapperImpl());
    const profile = profileMapper.mapFromDto(data);

    const validation = validateProfile(profile);
    if (!validation.isValid) {
      alert(validation.error);
      return;
    }

    axios
      .post("/api/profile", profile)
      .then(() => {
        navigation("/profile");
      })
      .catch((err) => {
        if (err.response) {
          const msg = "Status: " + err.response.status;
          if (err.response.data.cause) {
            alert(msg + ", cause: " + err.response.data.cause);
          } else if (err.response.data) {
            alert(msg + ", data:" + err.response.data);
          } else {
            alert(msg);
          }
        } else {
          alert(err.message);
        }
        console.log(err);
        return;
      });
  };

  const onUpdate = () => {
    console.log("trace: onUpdate");

    const data = getValues();
    const profileMapper = new ProfileMapperImpl(new FlowsMapperImpl());
    const profile = profileMapper.mapFromDto(data);

    const validation = validateProfile(profile);
    if (!validation.isValid) {
      alert(validation.error);
      return;
    }

    axios
      .put("/api/profile/" + profile.profileName, profile)
      .then(() => {
        navigation("/profile/" + profile.profileName);
      })
      .catch((err) => {
        if (err.response) {
          const msg = "Status: " + err.response.status;
          if (err.response.data.cause) {
            alert(msg + ", cause: " + err.response.data.cause);
          } else if (err.response.data) {
            alert(msg + ", data:" + err.response.data);
          } else {
            alert(msg);
          }
        }
      });
  };

    const formSubmitFn = isNewProfile ? onCreate : onUpdate;
    const formSubmitText = isNewProfile ? "CREATE" : "UPDATE";

    return (
    <Dashboard title="Profile" refreshAction={() => {}}>
        <form
            onSubmit={handleSubmit(formSubmitFn, (err: any) => {
                console.log("form error: ", err);
            })}
        >

            <ProfileFormBasic />

            <h3>Profile UE AMBR</h3>
            <ProfileFormUeAmbr />

            <ProfileFormSessions />

        <br />
        <Grid item xs={12}>
            <Button color="primary" variant="contained" type="submit" sx={{ m: 1 }}>
                {formSubmitText}
            </Button>
            </Grid>
        </form>
    </Dashboard>
    );
}

