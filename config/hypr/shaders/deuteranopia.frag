#version 300 es
precision highp float;
in vec2 v_texcoord;
uniform sampler2D tex;
out vec4 fragColor;
const mat3 seen = mat3(0.367322, 0.280085, -0.011820, 0.860646, 0.672501, 0.042940, -0.227968, 0.047413, 0.968881);
const mat3 shift = mat3(0.0, 0.7, 0.7, 0.0, 1.0, 0.0, 0.0, 0.0, 1.0);
void main() {
    vec4 pixel = texture(tex, v_texcoord);
    vec3 lost = pixel.rgb - seen * pixel.rgb;
    fragColor = vec4(clamp(pixel.rgb + shift * lost, 0.0, 1.0), pixel.a);
}
