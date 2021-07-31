import React, { useState, useEffect } from "react";
import { connect } from "react-redux";
import "react-dom";
import { Menu, SubMenu, MenuItem, MenuButton } from "@szhsin/react-menu";

import { fetchTags, fetchTagValues, updateTags } from "../redux/actions";
import history from "../util/history";
import "../util/prism";

function TagsBar({
  tags,
  fetchTags,
  fetchTagValues,
  updateTags,
  tagValuesLoading,
  labels,
}) {
  const [tagsValue, setTagsValue] = useState(
    new URL(window.location.href).searchParams.get("query") || "{}"
  );

  const loadTagValues = (tag) => {
    if (tags[tag] && !tags[tag].length && tagValuesLoading !== tag) {
      fetchTagValues(tag);
    }
  };

  const onTagsValueChange = (tag, tagValue) => {
    if (!tagsValue.includes(tag)) {
      setTagsValue(
        tagsValue.replace(
          "}",
          `${tagsValue === "{}" ? "" : ","}${tag}="${tagValue}"}`
        )
      );
    } else {
      const tagPairs = (tagsValue || "").replace(/[{}]/g, "").split(",");
      tagPairs.forEach((pair, i) => {
        if (pair.startsWith(tag)) {
          tagPairs[i] = `${tag}="${tagValue}"`;
        }
      });
      setTagsValue(`{${tagPairs.join(",")}}`);
    }
  };

  useEffect(() => {
    fetchTags();
  }, []);

  useEffect(() => {
    const url = new URL(window.location.href);
    Object.keys(tags).forEach((tag) => {
      if ((url.searchParams.get("query") || "").includes(tag)) {
        loadTagValues(tag);
      }
    });
  }, [tags]);

  useEffect(() => {
    const tagPairs = (tagsValue || "").replace(/[{}"]/g, "").split(",");
    const url = new URL(window.location.href);
    const tagsUpdater = [];
    tagPairs.forEach((pair) => {
      const [name, value] = pair.split("=");
      if (name && value) {
        tagsUpdater.push({ name, value });
      }
    });
    if (tagsValue) {
      url.searchParams.set("query", tagsValue);
      history.push(url.search);
      updateTags(tagsUpdater);
    }
    if (window.Prism) {
      window.Prism.highlightElement(
        document.getElementById("highlighting-content")
      );
    }
  }, [tagsValue]);

  return (
    <div className="tags-bar rc-menu-container--theme-dark">
      <Menu
        menuButton={<MenuButton>Select Tag</MenuButton>}
        theming="dark"
        keepMounted
      >
        {Object.keys(tags).map((tag) => (
          <SubMenu
            value={tag}
            key={tag}
            label={(e) => (
              <span
                className="tag-content"
                aria-hidden
                onClick={() => {
                  if (!tags[tag].length && tagValuesLoading !== tag)
                    loadTagValues(tag);
                }}
              >
                {tag}
              </span>
            )}
            className="active"
          >
            {tagValuesLoading === tag ? (
              <MenuItem>Loading...</MenuItem>
            ) : (
              tags[tag].map((tagValue) => (
                <MenuItem
                  key={tagValue}
                  value={tagValue}
                  onClick={(e) => onTagsValueChange(tag, e.value)}
                  className={
                    tagsValue.includes(`${tag}="${tagValue}"`) ? "active" : ""
                  }
                >
                  {tagValue}
                </MenuItem>
              ))
            )}
          </SubMenu>
        ))}
      </Menu>
      <div className="tags-query">
        <span className="tags-app-name">
          {labels && labels.find((label) => label.name === "__name__").value}
        </span>
        <pre className="tags-highlighted language-promql" aria-hidden="true">
          <code className="language-promql" id="highlighting-content">
            {tagsValue}
          </code>
        </pre>
        <input
          className="tags-input"
          type="text"
          value={tagsValue}
          onChange={(e) => setTagsValue(e.target.value)}
        />
      </div>
    </div>
  );
}

export default connect((state) => state, {
  fetchTags,
  fetchTagValues,
  updateTags,
})(TagsBar);
