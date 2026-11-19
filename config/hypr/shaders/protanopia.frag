#version 300 es
precision highp float;
in vec2 v_texcoord;
uniform sampler2D tex;
out vec4 fragColor;
const mat3 seen = mat3(0.152286, 0.114503, -0.003882, 1.052583, 0.786281, -0.048116, -0.204868, 0.099216, 1.051998);
const mat3 shift = mat3(0.0, 0.7, 0.7, 0.0, 1.0, 0.0, 0.0, 0.0, 1.0);
void main() {
    vec4 pixel = texture(tex, v_texcoord);
    vec3 lost = pixel.rgb - seen * pixel.rgb;
    fragColor = vec4(clamp(pixel.rgb + shift * lost, 0.0, 1.0), pixel.a);
}
